package service

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
	"strconv"
)

type StudentHandler struct{ repo repository.StudentRepository }

func NewStudentHandler(r repository.StudentRepository) *StudentHandler {
	return &StudentHandler{repo: r}
}
func (h *StudentHandler) List(c *fiber.Ctx) error {
	q := helper.ParseListQuery(c)
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	items, total, e := h.repo.FindAll(ctx, q)
	if e != nil {
		return helper.Fail(c, 500, "gagal mengambil daftar mahasiswa")
	}
	return helper.OKList(c, "daftar mahasiswa berhasil diambil", items, &model.Meta{Page: q.Page, Limit: q.Limit, Total: total, TotalPages: CountTotalPages(total, q.Limit)})
}
func (h *StudentHandler) Get(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, 400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	s, e := h.repo.FindByID(ctx, id)
	if errors.Is(e, repository.ErrNotFound) {
		return helper.Fail(c, 404, "mahasiswa tidak ditemukan")
	}
	if e != nil {
		return helper.Fail(c, 500, "gagal mengambil mahasiswa")
	}
	return helper.OK(c, "mahasiswa ditemukan", s)
}
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var r model.CreateStudentRequest
	if e := c.BodyParser(&r); e != nil {
		return helper.Fail(c, 400, "body harus berupa JSON yang valid")
	}
	if v := ValidateCreate(r); len(v) > 0 {
		return helper.FailValidation(c, v)
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	s, e := h.repo.Create(ctx, model.Student{NIM: r.NIM, Name: r.Name, Grade: r.Grade, IsActive: r.IsActive})
	if errors.Is(e, repository.ErrDuplicate) {
		return helper.Fail(c, 409, "NIM sudah terdaftar")
	}
	if e != nil {
		return helper.Fail(c, 500, "gagal menambahkan mahasiswa")
	}
	return helper.Created(c, "mahasiswa berhasil ditambahkan", s, "/api/v1/students/"+strconv.Itoa(s.ID))
}
func (h *StudentHandler) Replace(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, 400, "id harus berupa angka positif")
	}
	var r model.ReplaceStudentRequest
	if e := c.BodyParser(&r); e != nil {
		return helper.Fail(c, 400, "body harus berupa JSON yang valid")
	}
	if v := ValidateUpdate(r); len(v) > 0 {
		return helper.FailValidation(c, v)
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	s, e := h.repo.Update(ctx, model.Student{ID: id, NIM: r.NIM, Name: r.Name, Grade: r.Grade, IsActive: r.IsActive})
	if errors.Is(e, repository.ErrNotFound) {
		return helper.Fail(c, 404, "mahasiswa tidak ditemukan")
	}
	if errors.Is(e, repository.ErrDuplicate) {
		return helper.Fail(c, 409, "NIM sudah digunakan oleh mahasiswa lain")
	}
	if e != nil {
		return helper.Fail(c, 500, "gagal memperbarui mahasiswa")
	}
	return helper.OK(c, "data mahasiswa berhasil diganti seluruhnya", s)
}
func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, 400, "id harus berupa angka positif")
	}
	var r model.PatchStudentRequest
	if e := c.BodyParser(&r); e != nil {
		return helper.Fail(c, 400, "body harus berupa JSON yang valid")
	}
	if r.NIM == nil && r.Name == nil && r.Grade == nil && r.IsActive == nil {
		return helper.Fail(c, 400, "tidak ada field yang diubah")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	s, e := h.repo.FindByID(ctx, id)
	if errors.Is(e, repository.ErrNotFound) {
		return helper.Fail(c, 404, "mahasiswa tidak ditemukan")
	}
	if e != nil {
		return helper.Fail(c, 500, "gagal mengambil mahasiswa")
	}
	s, v := ApplyPatch(s, r)
	if len(v) > 0 {
		return helper.FailValidation(c, v)
	}
	s, e = h.repo.Update(ctx, s)
	if errors.Is(e, repository.ErrDuplicate) {
		return helper.Fail(c, 409, "NIM sudah digunakan oleh mahasiswa lain")
	}
	if e != nil {
		return helper.Fail(c, 500, "gagal memperbarui mahasiswa")
	}
	return helper.OK(c, "data mahasiswa berhasil diperbarui sebagian", s)
}
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, 400, "id harus berupa angka positif")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	e := h.repo.Delete(ctx, id)
	if errors.Is(e, repository.ErrNotFound) {
		return helper.Fail(c, 404, "mahasiswa tidak ditemukan")
	}
	if e != nil {
		return helper.Fail(c, 500, "gagal menghapus mahasiswa")
	}
	return helper.NoContent(c)
}
