package service

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/config"
	"latihan-fiber/helper"
)

// fakeCourseRepository adalah tiruan repository.CourseRepository.
// Dipakai untuk membuktikan service bisa diuji tanpa database.
type fakeCourseRepository struct {
	enrollErr error
	enrolled  bool
}

func (f *fakeCourseRepository) List(context.Context) ([]model.Course, error) {
	return []model.Course{}, nil
}

func (f *fakeCourseRepository) FindByID(_ context.Context, id int) (model.Course, error) {
	return model.Course{ID: id}, nil
}

func (f *fakeCourseRepository) IsEnrolled(_ context.Context, _, _ int) (bool, error) {
	return true, nil
}

func (f *fakeCourseRepository) ListByUser(context.Context, int) ([]model.Course, error) {
	return []model.Course{}, nil
}

func (f *fakeCourseRepository) Enroll(_ context.Context, _, _ int) (model.Course, error) {
	if f.enrollErr != nil {
		return model.Course{}, f.enrollErr
	}
	f.enrolled = true
	return model.Course{ID: 7, Code: "IF-101", Name: "Kalkulus", Credits: 3}, nil
}

func (f *fakeCourseRepository) Drop(_ context.Context, _, _ int) error { return nil }

func (f *fakeCourseRepository) Swap(_ context.Context, _, _, _ int) (model.Course, error) {
	return model.Course{}, nil
}

func (f *fakeCourseRepository) UpdateQuota(_ context.Context, _ int, _ model.UpdateQuotaRequest) (model.Course, error) {
	return model.Course{}, nil
}

// Fake di atas memenuhi kontrak repository.CourseRepository: service hanya
// mengenal interface, jadi bisa diuji tanpa database.
var _ repository.CourseRepository = (*fakeCourseRepository)(nil)

// TestCourseServiceEnrollWithFakeRepository memanggil handler Enroll dengan
// repository tiruan dan memastikan jalur sukses tetap 201 + isi JSON.
func TestCourseServiceEnrollWithFakeRepository(t *testing.T) {
	t.Parallel()

	fake := &fakeCourseRepository{}
	svc := CourseService{Repo: fake}

	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
	app.Post("/courses/:id/enroll", func(c *fiber.Ctx) error {
		c.Locals("user_id", 1)
		return svc.Enroll(c)
	})

	req := httptest.NewRequest("POST", "/courses/7/enroll", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, mau %d", resp.StatusCode, fiber.StatusCreated)
	}
	if !fake.enrolled {
		t.Fatal("repository tiruan tidak dipanggil")
	}

	var course model.Course
	if err := json.NewDecoder(resp.Body).Decode(&course); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if course.Code != "IF-101" {
		t.Fatalf("code = %q, mau IF-101", course.Code)
	}
}

// TestCourseServiceEnrollErrorMapping memastikan pemetaan error repository
// ke HTTP tetap sama setelah repository jadi interface.
func TestCourseServiceEnrollErrorMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		repoErr    error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "sudah terdaftar -> 409 ALREADY_ENROLLED",
			repoErr:    repository.ErrAlreadyEnrolled,
			wantStatus: fiber.StatusConflict,
			wantCode:   "ALREADY_ENROLLED",
		},
		{
			name:       "kuota penuh -> 409 CONFLICT",
			repoErr:    repository.ErrCourseFull,
			wantStatus: fiber.StatusConflict,
			wantCode:   helper.CodeConflict,
		},
		{
			name:       "batas SKS -> 409 CONFLICT",
			repoErr:    repository.ErrCreditLimit,
			wantStatus: fiber.StatusConflict,
			wantCode:   helper.CodeConflict,
		},
		{
			name:       "kelas tidak ada -> 404 NOT_FOUND",
			repoErr:    pgx.ErrNoRows,
			wantStatus: fiber.StatusNotFound,
			wantCode:   helper.CodeNotFound,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := CourseService{Repo: &fakeCourseRepository{enrollErr: tc.repoErr}}

			app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
			app.Post("/courses/:id/enroll", func(c *fiber.Ctx) error {
				c.Locals("user_id", 1)
				return svc.Enroll(c)
			})

			req := httptest.NewRequest("POST", "/courses/7/enroll", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, mau %d", resp.StatusCode, tc.wantStatus)
			}

			var body model.ErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Code != tc.wantCode {
				t.Fatalf("code = %q, mau %q", body.Code, tc.wantCode)
			}
		})
	}
}
