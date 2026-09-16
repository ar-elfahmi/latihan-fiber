package service

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// PrestasiService adalah struct konkret untuk tabel prestasi.
type PrestasiService struct {
	Repo repository.PrestasiRepository
}

// sortColumns whitelist untuk ORDER BY tabel prestasi.
var prestasiSortColumns = map[string]bool{
	"id": true, "student_id": true, "keterangan": true, "juara": true, "created_at": true,
}

// List menampilkan daftar prestasi (paginasi + filter).
func (s PrestasiService) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	q := model.ListQuery{
		Page: page, Limit: limit,
		Search: c.Query("search"),
		Sort:   c.Query("sort", "id"),
		Order:  c.Query("order", "asc"),
	}
	if !prestasiSortColumns[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}

	items, total, err := s.Repo.GetAll(c.Context(), q)
	if err != nil {
		return helper.Error(c, 500, "gagal mengambil daftar prestasi")
	}
	return c.JSON(fiber.Map{
		"data": items,
		"meta": fiber.Map{
			"page": q.Page, "limit": q.Limit, "total": total,
			"total_pages": CountTotalPages(total, q.Limit),
		},
	})
}

// GetByStudent menampilkan seluruh prestasi milik seorang mahasiswa.
func (s PrestasiService) GetByStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Error(c, 400, "id mahasiswa harus berupa angka positif")
	}
	items, err := s.Repo.GetByStudentID(c.Context(), id)
	if err != nil {
		return helper.Error(c, 500, "gagal mengambil prestasi mahasiswa")
	}
	return c.JSON(items)
}

// Create menambahkan prestasi baru untuk seorang mahasiswa.
// student_id selalu diambil dari :id, bukan dari body.
func (s PrestasiService) Create(c *fiber.Ctx) error {
	studentID, err := strconv.Atoi(c.Params("id"))
	if err != nil || studentID < 1 {
		return helper.Error(c, 400, "id mahasiswa harus berupa angka positif")
	}

	var req model.CreatePrestasiRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.Error(c, 400, "body harus berupa JSON yang valid")
	}
	if errs := ValidatePrestasi(req); len(errs) > 0 {
		return c.Status(422).JSON(fiber.Map{"error": "validasi gagal", "fields": errs})
	}

	created, err := s.Repo.Create(c.Context(), model.Prestasi{
		StudentID: studentID, Keterangan: req.Keterangan, Juara: req.Juara,
	})
	if err != nil {
		if err == repository.ErrStudentNotFound {
			return helper.Error(c, 404, "mahasiswa tidak ditemukan")
		}
		return helper.Error(c, 500, "gagal menambahkan prestasi")
	}

	c.Set("Location", "/api/v1/students/"+strconv.Itoa(studentID)+"/prestasi")
	return c.Status(201).JSON(created)
}
