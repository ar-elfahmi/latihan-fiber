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

func (s CourseService) List(c *fiber.Ctx) error {
	courses, err := s.Repo.List(c.UserContext())
	if err != nil {
		return helper.Internal(err)
	}
	return c.JSON(courses)
}

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
	return c.JSON(model.WebResponse{
		Success: true,
		Message: "KRS berhasil diambil",
		Data: model.MyCourses{
			Courses:      courses,
			TotalCredits: total,
			MaxCredits:   MaxCreditsPerSemester,
		},
	})
}

func (s CourseService) Detail(c *fiber.Ctx) error {
	id, err := courseID(c)
	if err != nil {
		return err
	}
	course, err := s.Repo.FindByID(c.UserContext(), id)
	if err != nil {
		return translateCourseError(err)
	}
	return c.JSON(model.WebResponse{
		Success: true,
		Message: "detail kelas berhasil diambil",
		Data:    course,
	})
}

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
