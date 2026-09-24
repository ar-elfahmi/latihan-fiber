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
	Repo    repository.PrestasiRepository
	Student repository.StudentRepository
}

// prestasiSortColumns whitelist untuk ORDER BY tabel prestasi.
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
		Page:  page,
		Limit: limit,
	}
	if !prestasiSortColumns[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}

	items, total, err := s.Repo.GetAll(c.Context(), q)
	if err != nil {
		return helper.Internal(err)
	}
	return c.JSON(fiber.Map{
		"data":        items,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": CountTotalPages(total, limit),
		"sort":        q.Sort,
		"order":       q.Order,
	})
}

// GetByStudent menampilkan seluruh prestasi milik seorang mahasiswa.
// Modul 6: kalau user hanya punya prestasi:read:self -> cek student adalah miliknya.
func (s PrestasiService) GetByStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return helper.BadRequest("id tidak valid")
	}
	perms := CurrentPerms(c)
	if !perms["prestasi:read:any"] {
		// Hanya :self (atau tak punya). Wajib cek pemilikan student-nya.
		st, err := s.Student.GetByID(c.Context(), id)
		if err != nil {
			return translateError(err, "mahasiswa")
		}
		owner := 0
		if st.OwnerID != nil {
			owner = *st.OwnerID
		}
		if !RequireOwnerOrAdmin(CurrentUserID(c), owner) {
			return helper.Forbidden("anda hanya boleh melihat prestasi milik sendiri")
		}
	}
	items, err := s.Repo.GetByStudentID(c.Context(), id)
	if err != nil {
		return helper.Internal(err)
	}
	return c.JSON(fiber.Map{"data": items})
}

// Create menambahkan prestasi baru untuk seorang mahasiswa.
// student_id selalu diambil dari :id, bukan dari body.
// Modul 6: kalau user hanya punya student:create tanpa akses admin/staff,
// wajib cek student adalah miliknya.
func (s PrestasiService) Create(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return helper.BadRequest("id tidak valid")
	}

	var req model.CreatePrestasiRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := ValidatePrestasi(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	// Modul 6: hanya admin/staff yang boleh menambah prestasi untuk student
	// lain. User biasa yang 'menciptakan diri sendiri' ditolak di sini.
	perms := CurrentPerms(c)
	if !perms["student:read:any"] {
		st, err := s.Student.GetByID(c.Context(), id)
		if err != nil {
			return translateError(err, "mahasiswa")
		}
		owner := 0
		if st.OwnerID != nil {
			owner = *st.OwnerID
		}
		if !RequireOwnerOrAdmin(CurrentUserID(c), owner) {
			return helper.Forbidden("hanya admin/staff yang boleh menambah prestasi untuk mahasiswa lain")
		}
	}

	created, err := s.Repo.Create(c.Context(), model.Prestasi{
		StudentID:  id,
		Juara:      req.Juara,
		Keterangan: req.Keterangan,
	})
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	return c.Status(201).JSON(created)
}
