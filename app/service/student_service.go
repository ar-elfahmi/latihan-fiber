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
	q := model.ListQuery{
		Page:  c.QueryInt("page", 1),
		Limit: c.QueryInt("limit", 10),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 50 {
		q.Limit = 10
	}
	q.Search = strings.TrimSpace(c.Query("search"))
	q.Sort = c.Query("sort")
	q.Order = c.Query("order")
	if v := c.Query("is_active"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			q.IsActive = &b
		}
	}
	return q
}

// idParam mengambil parameter :id dari URL, mengembalikan -1 bila bukan angka.
func idParam(c *fiber.Ctx) int {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return -1
	}
	return id
}

// List menampilkan daftar mahasiswa dengan paginasi.
func (s StudentService) List(c *fiber.Ctx) error {
	q := fromQueryString(c)
	if !sortColumns[q.Sort] {
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
		"page":        q.Page,
		"limit":       q.Limit,
		"total":       total,
		"total_pages": CountTotalPages(total, q.Limit),
		"sort":        q.Sort,
		"order":       q.Order,
	})
}

// Get menampilkan satu mahasiswa berdasarkan :id.
// Modul 6: kalau user hanya punya permission :self, cek kepemilikan.
func (s StudentService) Get(c *fiber.Ctx) error {
	id := idParam(c)
	if id <= 0 {
		return helper.BadRequest("id tidak valid")
	}
	perms := helper.PermissionSet(CurrentPerms(c))
	st, err := s.Repo.GetByID(c.Context(), id)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	me := model.AuthUser{ID: CurrentUserID(c)}
	owner := 0
	if st.OwnerID != nil {
		owner = *st.OwnerID
	}
	if !CanAccessStudent(me, owner, &perms, "student:read:any") {
		return helper.Forbidden("anda hanya boleh melihat data milik sendiri")
	}
	return c.JSON(st)
}

// Create menambahkan mahasiswa baru. Validasi dipanggil dari rules.
// Modul 6: owner_id SELALU diisi dari identitas pemanggil (token),
// tidak pernah dari body. Struct CreateStudentRequest memang tidak
// punya field owner_id, jadi sekalipun client mengirim "owner_id": 1
// di body, encoding/json mengabaikannya (anti mass-assignment).
// Dengan begitu tidak ada yang bisa mengaku jadi pemilik orang lain.
// Permission student:create sendiri dicek di middleware route.
func (s StudentService) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.Validation(errs)
	}
	actor := CurrentUserID(c)
	var owner *int
	if actor > 0 {
		owner = &actor
	}
	st := model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  owner,
	}
	created, err := s.Repo.Create(c.Context(), st)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	return c.Status(201).JSON(created)
}

// Replace (PUT) mengganti seluruh field mahasiswa.
// Modul 6: pemilik hanya boleh edit miliknya (kalau :self permission).
func (s StudentService) Replace(c *fiber.Ctx) error {
	id := idParam(c)
	if id <= 0 {
		return helper.BadRequest("id tidak valid")
	}
	if err := enforceStudentWrite(c, s, id, "student:update:any", "student:update:self"); err != nil {
		return err
	}
	var req model.ReplaceStudentRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := ValidateReplace(req); len(errs) > 0 {
		return helper.Validation(errs)
	}
	st := model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}
	updated, err := s.Repo.Replace(c.Context(), st)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	return c.JSON(updated)
}

// Patch mengubah sebagian field mahasiswa (hanya field yang dikirim).
func (s StudentService) Patch(c *fiber.Ctx) error {
	id := idParam(c)
	if id <= 0 {
		return helper.BadRequest("id tidak valid")
	}
	if err := enforceStudentWrite(c, s, id, "student:update:any", "student:update:self"); err != nil {
		return err
	}
	var req model.PatchStudentRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if IsEmptyPatch(req) {
		return helper.Validation(map[string]string{"body": "tidak ada field yang dikirim"})
	}
	st, err := s.Repo.GetByID(c.Context(), id)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	updated, errs := ApplyPatch(st, req)
	if len(errs) > 0 {
		return helper.Validation(errs)
	}
	saved, err := s.Repo.Replace(c.Context(), updated)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	return c.JSON(saved)
}

// Delete menghapus mahasiswa. Sukses mengembalikan 204 tanpa body.
func (s StudentService) Delete(c *fiber.Ctx) error {
	id := idParam(c)
	if id <= 0 {
		return helper.BadRequest("id tidak valid")
	}
	if err := s.Repo.Delete(c.Context(), id); err != nil {
		return translateError(err, "mahasiswa")
	}
	return c.SendStatus(204)
}

// enforceStudentWrite memutuskan apakah user boleh menulis ke record
// student (PUT/PATCH). Seluruh keputusan lewat CanAccessStudent:
// pemegang :any boleh; selain itu harus pemilik record. Record tanpa
// owner hanya bisa diubah pemegang :any.
// Pemeriksaan grup (any/self) diulang di sini sebagai lapis kedua
// (defense in depth) untuk berjaga bila handler dipanggil di luar route.
func enforceStudentWrite(c *fiber.Ctx, s StudentService, id int, anyPerm, selfPerm string) error {
	perms := CurrentPerms(c)
	if !HasAnyPermissionFromGroups(perms, []string{anyPerm, selfPerm}) {
		return helper.Forbidden("tidak punya hak untuk aksi ini")
	}
	st, err := s.Repo.GetByID(c.Context(), id)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	owner := 0
	if st.OwnerID != nil {
		owner = *st.OwnerID
	}
	me := model.AuthUser{ID: CurrentUserID(c)}
	ps := helper.PermissionSet(perms)
	if !CanAccessStudent(me, owner, &ps, anyPerm) {
		return helper.Forbidden("anda hanya boleh mengubah data milik sendiri")
	}
	return nil
}
