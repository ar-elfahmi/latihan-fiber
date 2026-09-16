package service

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// StudentService adalah struct konkret (tanpa interface) sesuai gaya
// landing-page. Ia memegang repository langsung lewat field Repo.
type StudentService struct {
	Repo repository.StudentRepository
}

// fromQueryString membaca ListQuery dari query string request.
// Dipakai oleh GET /students.
func fromQueryString(c *fiber.Ctx) model.ListQuery {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	q := model.ListQuery{
		Page:   page,
		Limit:  limit,
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}
	if !sortColumns[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}
	return q
}

// idParam mengambil parameter :id dari URL, mengembalikan -1 bila bukan angka.
func idParam(c *fiber.Ctx) int {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return -1
	}
	return id
}

// List menampilkan daftar mahasiswa dengan paginasi.
func (s StudentService) List(c *fiber.Ctx) error {
	q := fromQueryString(c)
	items, total, err := s.Repo.GetAll(c.Context(), q)
	if err != nil {
		return helper.Error(c, 500, "gagal mengambil daftar mahasiswa")
	}
	return c.JSON(fiber.Map{
		"data": items,
		"meta": fiber.Map{
			"page":        q.Page,
			"limit":       q.Limit,
			"total":       total,
			"total_pages": CountTotalPages(total, q.Limit),
		},
	})
}

// Get menampilkan satu mahasiswa berdasarkan :id.
func (s StudentService) Get(c *fiber.Ctx) error {
	id := idParam(c)
	if id < 1 {
		return helper.Error(c, 400, "id harus berupa angka positif")
	}
	student, err := s.Repo.GetByID(c.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			return helper.Error(c, 404, "mahasiswa tidak ditemukan")
		}
		return helper.Error(c, 500, "gagal mengambil mahasiswa")
	}
	return c.JSON(student)
}

// Create menambahkan mahasiswa baru. Validasi dipanggil dari rules.
func (s StudentService) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.Error(c, 400, "body harus berupa JSON yang valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := ValidateCreate(req); len(errs) > 0 {
		return c.Status(422).JSON(fiber.Map{"error": "validasi gagal", "fields": errs})
	}

	student, err := s.Repo.Create(c.Context(), model.Student{
		NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: req.IsActive,
	})
	if err != nil {
		return mapStudentError(c, err, "mahasiswa tidak ditemukan",
			"NIM sudah terdaftar", "gagal menambahkan mahasiswa")
	}

	c.Set("Location", "/api/v1/students/"+strconv.Itoa(student.ID))
	return c.Status(201).JSON(student)
}

// Replace (PUT) mengganti seluruh field mahasiswa.
func (s StudentService) Replace(c *fiber.Ctx) error {
	id := idParam(c)
	if id < 1 {
		return helper.Error(c, 400, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.Error(c, 400, "body harus berupa JSON yang valid")
	}

	if errs := ValidateReplace(req); len(errs) > 0 {
		return c.Status(422).JSON(fiber.Map{"error": "validasi gagal", "fields": errs})
	}

	student, err := s.Repo.Replace(c.Context(), model.Student{
		ID: id, NIM: strings.TrimSpace(req.NIM), Name: strings.TrimSpace(req.Name),
		Grade: req.Grade, IsActive: req.IsActive,
	})
	if err != nil {
		return mapStudentError(c, err, "mahasiswa tidak ditemukan",
			"NIM sudah digunakan oleh mahasiswa lain", "gagal memperbarui mahasiswa")
	}
	return c.JSON(student)
}

// Patch mengubah sebagian field mahasiswa (hanya field yang dikirim).
func (s StudentService) Patch(c *fiber.Ctx) error {
	id := idParam(c)
	if id < 1 {
		return helper.Error(c, 400, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.Error(c, 400, "body harus berupa JSON yang valid")
	}
	if IsEmptyPatch(req) {
		return helper.Error(c, 400, "tidak ada field yang diubah")
	}

	current, err := s.Repo.GetByID(c.Context(), id)
	if err != nil {
		return mapStudentError(c, err, "mahasiswa tidak ditemukan",
			"", "gagal mengambil mahasiswa")
	}
	updated, errs := ApplyPatch(current, req)
	if len(errs) > 0 {
		return c.Status(422).JSON(fiber.Map{"error": "validasi gagal", "fields": errs})
	}

	student, err := s.Repo.Replace(c.Context(), updated)
	if err != nil {
		return mapStudentError(c, err, "mahasiswa tidak ditemukan",
			"NIM sudah digunakan oleh mahasiswa lain", "gagal memperbarui mahasiswa")
	}
	return c.JSON(student)
}

// Delete menghapus mahasiswa. Sukses mengembalikan 204 tanpa body.
func (s StudentService) Delete(c *fiber.Ctx) error {
	id := idParam(c)
	if id < 1 {
		return helper.Error(c, 400, "id harus berupa angka positif")
	}
	err := s.Repo.Delete(c.Context(), id)
	if err != nil {
		return mapStudentError(c, err, "mahasiswa tidak ditemukan",
			"", "gagal menghapus mahasiswa")
	}
	return c.SendStatus(204)
}

// mapStudentError menerjemahkan error repository ke status HTTP.
// Versi inline (tanpa file translateError terpisah) selaras dengan landing-page.
func mapStudentError(c *fiber.Ctx, err error, notFoundMsg, duplicateMsg, generalMsg string) error {
	switch {
	case err == repository.ErrNotFound:
		return helper.Error(c, 404, notFoundMsg)
	case duplicateMsg != "" && err == repository.ErrDuplicate:
		return helper.Error(c, 409, duplicateMsg)
	default:
		return helper.Error(c, 500, generalMsg)
	}
}
