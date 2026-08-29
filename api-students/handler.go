package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
)

type StudentHandler struct {
	repo repository.StudentRepository
}

func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{
		repo: repo,
	}
}

func (h *StudentHandler) List(c *fiber.Ctx) error {
	q := parseListQuery(c)

	ctx, cancel := reqCtx(c)
	defer cancel()

	students, total, err := h.repo.FindAll(ctx, q)
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar mahasiswa")
	}

	totalPages := (total + q.Limit - 1) / q.Limit

	return okList(
		c,
		"daftar mahasiswa berhasil diambil",
		students,
		&model.Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	)
}

func (h *StudentHandler) Get(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	ctx, cancel := reqCtx(c)
	defer cancel()

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		return fail(c, fiber.StatusInternalServerError, "gagal mengambil mahasiswa")
	}

	return ok(c, "mahasiswa ditemukan", student)
}

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 4.0 {
		errs["grade"] = "harus bernilai antara 0.0 - 4.0"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student := model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	ctx, cancel := reqCtx(c)
	defer cancel()

	student, err := h.repo.Create(ctx, student)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}

		return fail(c, fiber.StatusInternalServerError, "gagal menambahkan mahasiswa")
	}

	return created(
		c,
		"mahasiswa berhasil ditambahkan",
		student,
		"/api/v1/students/"+strconv.Itoa(student.ID),
	)
}

func (h *StudentHandler) Replace(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}

	if req.Name == "" {
		errs["name"] = "wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 4.0 {
		errs["grade"] = "harus bernilai antara 0.0 - 4.0"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student := model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	ctx, cancel := reqCtx(c)
	defer cancel()

	student, err := h.repo.Update(ctx, student)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(c, fiber.StatusConflict, "NIM sudah digunakan oleh mahasiswa lain")
		}

		return fail(c, fiber.StatusInternalServerError, "gagal memperbarui mahasiswa")
	}

	return ok(c, "data mahasiswa berhasil diganti seluruhnya", student)
}

func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	ctx, cancel := reqCtx(c)
	defer cancel()

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		return fail(c, fiber.StatusInternalServerError, "gagal mengambil mahasiswa")
	}

	errs := map[string]string{}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)

		if nim == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			student.NIM = nim
		}
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			student.Name = name
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 4.0 {
			errs["grade"] = "harus bernilai antara 0.0 - 4.0"
		} else {
			student.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student, err = h.repo.Update(ctx, student)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(c, fiber.StatusConflict, "NIM sudah digunakan oleh mahasiswa lain")
		}

		return fail(c, fiber.StatusInternalServerError, "gagal memperbarui mahasiswa")
	}

	return ok(c, "data mahasiswa berhasil diperbarui sebagian", student)
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	ctx, cancel := reqCtx(c)
	defer cancel()

	if err := h.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		return fail(c, fiber.StatusInternalServerError, "gagal menghapus mahasiswa")
	}

	return noContent(c)
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}
