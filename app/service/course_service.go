package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type CourseService struct {
	Repo repository.CourseRepository
}

// List mengembalikan seluruh kelas yang tersedia.
//
// @Summary		Daftar kelas
// @Description	Mengembalikan seluruh kelas yang tersedia beserta kuota dan jumlah peserta.
// @Tags			Courses
// @Produce		json
// @Success		200	{array}		model.Course
// @Failure		500	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/courses [get]
func (s CourseService) List(c *fiber.Ctx) error {
	courses, err := s.Repo.List(c.UserContext())
	if err != nil {
		return helper.Internal(err)
	}
	return c.JSON(courses)
}

// Mine mengembalikan KRS milik user yang sedang login.
//
// @Summary		Daftar KRS saya
// @Description	Mengembalikan kelas yang diambil user beserta total SKS dan batas maksimum SKS per semester.
// @Tags			Courses
// @Produce		json
// @Success		200	{object}	model.MyCoursesResponse
// @Failure		401	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/courses/mine [get]
func (s CourseService) Mine(c *fiber.Ctx) error {
	courses, err := s.Repo.ListByUser(c.UserContext(), CurrentUserID(c))
	if err != nil {
		return translateCourseError(err)
	}
	if courses == nil {
		courses = []model.Course{}
	}
	total := 0
	for _, course := range courses {
		total += course.Credits
	}
	return c.JSON(model.MyCoursesResponse{
		Success: true,
		Message: "KRS berhasil diambil",
		Data: model.MyCourses{
			Courses:      courses,
			TotalCredits: total,
			MaxCredits:   MaxCreditsPerSemester,
		},
	})
}

// Detail mengembalikan detail satu kelas.
//
// @Summary		Detail satu kelas
// @Description	Mengembalikan detail satu kelas: kode, nama, SKS, kuota, dan jumlah peserta.
// @Tags			Courses
// @Produce		json
// @Param			id	path		int	true	"ID kelas"
// @Success		200	{object}	model.CourseResponse
// @Failure		400	{object}	model.ErrorResponse
// @Failure		404	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/courses/{id} [get]
func (s CourseService) Detail(c *fiber.Ctx) error {
	id, err := courseID(c)
	if err != nil {
		return err
	}
	course, err := s.Repo.FindByID(c.UserContext(), id)
	if err != nil {
		return translateCourseError(err)
	}
	return c.JSON(model.CourseResponse{
		Success: true,
		Message: "detail kelas berhasil diambil",
		Data:    course,
	})
}

// Enroll mendaftarkan user ke satu kelas.
//
// @Summary		Ambil kelas (KRS)
// @Description	Mendaftarkan user ke kelas. Gagal bila kuota penuh (409), sudah terdaftar (409), atau total SKS melebihi batas (409).
// @Tags			Courses
// @Produce		json
// @Param			id	path		int	true	"ID kelas"
// @Success		201	{object}	model.Course
// @Failure		401	{object}	model.ErrorResponse
// @Failure		404	{object}	model.ErrorResponse
// @Failure		409	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/courses/{id}/enroll [post]
func (s CourseService) Enroll(c *fiber.Ctx) error {
	id, err := courseID(c)
	if err != nil {
		return err
	}
	course, err := s.Repo.Enroll(c.UserContext(), CurrentUserID(c), id)
	if err != nil {
		return translateCourseError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(course)
}

// Drop membatalkan pengambilan satu kelas.
//
// @Summary		Batalkan ambil kelas (drop)
// @Description	Membatalkan pengambilan satu kelas dari KRS user.
// @Tags			Courses
// @Produce		json
// @Param			id	path		int	true	"ID kelas"
// @Success		204
// @Failure		400	{object}	model.ErrorResponse
// @Failure		404	{object}	model.ErrorResponse
// @Failure		409	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/courses/{id}/enroll [delete]
func (s CourseService) Drop(c *fiber.Ctx) error {
	id, err := courseID(c)
	if err != nil {
		return err
	}
	if err := s.Repo.Drop(c.UserContext(), CurrentUserID(c), id); err != nil {
		return translateCourseError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Swap menukar satu kelas dengan kelas lain dalam satu transaksi.
//
// @Summary		Tukar kelas (swap)
// @Description	Menukar kelas asal dengan kelas tujuan. Kuota, SKS, dan status peserta dijaga dalam satu transaksi.
// @Tags			Courses
// @Accept			json
// @Produce		json
// @Param			request	body		model.SwapRequest	true	"Kelas asal dan tujuan"
// @Success		200	{object}	model.Course
// @Failure		400	{object}	model.ErrorResponse
// @Failure		404	{object}	model.ErrorResponse
// @Failure		409	{object}	model.ErrorResponse
// @Failure		422	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/enrollments/swap [post]
func (s CourseService) Swap(c *fiber.Ctx) error {
	var req model.SwapRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	if req.FromCourseID == req.ToCourseID {
		return helper.Validation(map[string]string{"to_course_id": "kelas tujuan harus berbeda dari kelas asal"})
	}
	course, err := s.Repo.Swap(c.UserContext(), CurrentUserID(c), req.FromCourseID, req.ToCourseID)
	if err != nil {
		return translateCourseError(err)
	}
	return c.JSON(course)
}

// UpdateQuota mengubah kuota satu kelas (admin).
//
// @Summary		Ubah kuota kelas
// @Description	Mengubah kuota peserta satu kelas. Memerlukan version untuk optimistic locking.
// @Tags			Courses
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"ID kelas"
// @Param			request	body		model.UpdateQuotaRequest	true	"Kuota baru dan versi"
// @Success		200	{object}	model.Course
// @Failure		400	{object}	model.ErrorResponse
// @Failure		404	{object}	model.ErrorResponse
// @Failure		409	{object}	model.ErrorResponse
// @Failure		422	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/courses/{id} [patch]
func (s CourseService) UpdateQuota(c *fiber.Ctx) error {
	id, err := courseID(c)
	if err != nil {
		return err
	}
	var req model.UpdateQuotaRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	course, err := s.Repo.UpdateQuota(c.UserContext(), id, req)
	if err != nil {
		return translateCourseError(err)
	}
	return c.JSON(course)
}

func courseID(c *fiber.Ctx) (int, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil || id < 1 {
		return 0, helper.BadRequest("id kelas harus berupa angka positif")
	}
	return int(id), nil
}

func translateCourseError(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return helper.NotFound("kelas")
	case errors.Is(err, repository.ErrAlreadyEnrolled):
		return &helper.AppError{Status: fiber.StatusConflict, Code: "ALREADY_ENROLLED", Message: err.Error()}
	case errors.Is(err, repository.ErrNotEnrolled),
		errors.Is(err, repository.ErrCourseFull),
		errors.Is(err, repository.ErrCreditLimit),
		errors.Is(err, repository.ErrQuotaConflict):
		return helper.Conflict(err.Error())
	default:
		return helper.Internal(err)
	}
}
