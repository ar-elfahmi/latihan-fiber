package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// StudentService memegang dua tanggung jawab sekaligus pada struktur baku
// mata kuliah ini: menerima *fiber.Ctx (peran controller) dan menjalankan
// business rules (peran use case). Business rules-nya sendiri berada di
// student_rules.go, file terpisah yang tidak menyentuh Fiber sama sekali.
type StudentService struct {
	repo repository.StudentRepository
}

// NewStudentService menerima INTERFACE, bukan struct konkret.
func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	items, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError,
			"gagal mengambil daftar mahasiswa")
	}

	return helper.SuccessList(c, "daftar mahasiswa berhasil diambil", items, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err,
			"mahasiswa tidak ditemukan", "", "gagal mengambil mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest,
			"body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	// Business rulesnya dipanggil, bukan ditulis ulang di sini.
	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	newStudent, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})
	if err != nil {
		return translateError(c, err,
			"mahasiswa tidak ditemukan", "NIM sudah terdaftar",
			"gagal menambahkan mahasiswa")
	}

	return helper.Created(c, "mahasiswa berhasil ditambahkan", newStudent,
		"/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest,
			"body harus berupa JSON yang valid")
	}

	if errs := ValidateUpdate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})
	if err != nil {
		return translateError(c, err,
			"mahasiswa tidak ditemukan", "NIM sudah digunakan oleh mahasiswa lain",
			"gagal memperbarui mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diganti seluruhnya", result)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest,
			"body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err,
			"mahasiswa tidak ditemukan", "", "gagal mengambil mahasiswa")
	}

	updated, errs := ApplyPatch(current, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(c, err,
			"mahasiswa tidak ditemukan", "NIM sudah digunakan oleh mahasiswa lain",
			"gagal memperbarui mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diperbarui sebagian", result)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err,
			"mahasiswa tidak ditemukan", "", "gagal menghapus mahasiswa")
	}

	return helper.NoContent(c)
}

// translateError memetakan error milik repository menjadi status HTTP.
// Pesan untuk tidak-ditemukan dan duplikat bisa berbeda per endpoint,
// karena itu keduanya dikirim sebagai argumen.
func translateError(
	c *fiber.Ctx, err error, notFoundMsg, duplicateMsg, generalMsg string,
) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, notFoundMsg)
	case duplicateMsg != "" && errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, duplicateMsg)
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, generalMsg)
	}
}
