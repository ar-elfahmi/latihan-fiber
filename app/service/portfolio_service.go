package service

import (
	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
)

type PortfolioService struct {
	Repo        repository.PortfolioRepository
	StudentRepo repository.StudentRepository // untuk cek mahasiswa ada di PostgreSQL
}

// authorizeStudent memastikan pemanggil boleh mengakses data mahasiswa id.
//
// Lapis halus (data): pemohon yang hanya punya permission :self hanya boleh
// menyentuh baris yang owner_id-nya dia sendiri. CanAccessStudent sudah
// menutup jalur ":any" (admin/staff) dan fail closed untuk owner_id NULL.
//
// Lapis kasar (permission) ada di route.go: middleware.RequirePermission.
func (s PortfolioService) authorizeStudent(c *fiber.Ctx, st model.Student, anyPerm string) error {
	perms := helper.PermissionSet(CurrentPerms(c))
	me := model.AuthUser{ID: CurrentUserID(c)}
	owner := 0
	if st.OwnerID != nil {
		owner = *st.OwnerID
	}
	if !CanAccessStudent(me, owner, &perms, anyPerm) {
		return helper.Forbidden("anda hanya boleh mengakses data milik sendiri")
	}
	return nil
}

func (s PortfolioService) Get(c *fiber.Ctx) error {
	id := idParam(c) // fungsi ini sudah ada di student_service.go
	if id <= 0 {
		return helper.BadRequest("id tidak valid")
	}
	// Cek mahasiswa ada di PostgreSQL
	st, err := s.StudentRepo.GetByID(c.Context(), id)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	// Cek kepemilikan (permission student:read:any/:self)
	if err := s.authorizeStudent(c, st, "student:read:any"); err != nil {
		return err
	}
	// Ambil portofolio dari MongoDB
	portfolio, err := s.Repo.FindByStudentID(c.Context(), id)
	if err != nil {
		return translateError(err, "portofolio")
	}
	// kirim respon
	return c.JSON(portfolio)
}

func (s PortfolioService) Upsert(c *fiber.Ctx) error {
	id := idParam(c)
	if id <= 0 {
		return helper.BadRequest("id tidak valid")
	}

	// 1. Cek mahasiswa ada di PostgreSQL (sama seperti Get)
	st, err := s.StudentRepo.GetByID(c.Context(), id)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	// Cek kepemilikan (permission student:update:any/:self)
	if err := s.authorizeStudent(c, st, "student:update:any"); err != nil {
		return err
	}
	// 2. Baca body request
	var req model.UpsertPortfolioRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	// 3. Validasi
	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	// 4. Panggil repository Upsert
	certs := make([]model.Certificate, 0, len(req.Certificates))
	for _, cert := range req.Certificates { // ganti c → cert
		certs = append(certs, model.Certificate{
			Name:   cert.Name,
			Issuer: cert.Issuer,
			Date:   cert.Date,
		})
	}

	sp := model.Portfolio{
		StudentID:    id,
		Headline:     req.Headline,
		Bio:          req.Bio,
		Certificates: certs,
	}

	saved, err := s.Repo.Upsert(c.Context(), sp)
	if err != nil {
		return translateError(err, "portofolio")
	}

	// 5. Kirim response 200
	return c.JSON(saved)
}

func (s PortfolioService) AddProject(c *fiber.Ctx) error {
	id := idParam(c)
	if id <= 0 {
		return helper.BadRequest("id tidak valid")
	}

	// 1. Cek mahasiswa ada di PostgreSQL
	st, err := s.StudentRepo.GetByID(c.Context(), id)
	if err != nil {
		return translateError(err, "mahasiswa")
	}
	// Cek kepemilikan (permission student:update:any/:self)
	if err := s.authorizeStudent(c, st, "student:update:any"); err != nil {
		return err
	}

	// 2. Baca dan validasi body
	var req model.AddProjectRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	// 3. Panggil repository AddProject
	spj := model.Project{
		Title:        req.Title,
		Technologies: req.Technologies,
		URL:          req.URL,
		Year:         req.Year,
	}
	created, err := s.Repo.AddProject(c.Context(), id, spj)
	if err != nil {
		return translateError(err, "proyek")
	}
	// 4. Kirim response 201
	return c.Status(201).JSON(created)
}

func (s PortfolioService) DeleteProject(c *fiber.Ctx) error {
	id := idParam(c)
	if id <= 0 {
		return helper.BadRequest("id tidak valid")
	}

	projectID := c.Params("projectId")
	if projectID == "" {
		return helper.BadRequest("project id tidak valid")
	}

	// 1. Cek mahasiswa ada di PostgreSQL
	st, err := s.StudentRepo.GetByID(c.Context(), id)
	if err != nil {
		return translateError(err, "proyek")
	}
	// Cek kepemilikan (permission student:update:any/:self)
	if err := s.authorizeStudent(c, st, "student:update:any"); err != nil {
		return err
	}
	// 2. Panggil repository DeleteProject
	if err := s.Repo.DeleteProject(c.Context(), id, projectID); err != nil {
		return translateError(err, "proyek")
	}
	// 3. Kirim response 204 (no content — sama seperti Delete di student_service.go)
	return c.SendStatus(204)
}

func (s PortfolioService) StatsTechnologies(c *fiber.Ctx) error {
	// 1. Panggil repository
	stats, err := s.Repo.StatsTechnologies(c.Context())
	if err != nil {
		return helper.Internal(err)
	}
	// 2. Kirim response
	return c.JSON(stats)
}
