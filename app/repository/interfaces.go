package repository

import (
	"context"
	"time"

	"latihan-fiber/app/model"
)

// Repository adalah kontrak yang dipakai service. Service hanya mengenal
// interface di bawah; implementasi PostgreSQL (prefix "Postgres") tinggal
// satu dari banyak kemungkinan, sehingga test bisa memakai fake.
//
// Method set setiap interface = hanya method yang benar-benar dipakai
// service/worker/relay. Method yang tidak dipakai tidak masuk interface.

// CourseRepository adalah kontrak repository tabel courses.
type CourseRepository interface {
	List(ctx context.Context) ([]model.Course, error)
	Enroll(ctx context.Context, userID, courseID int) (model.Course, error)
	Drop(ctx context.Context, userID, courseID int) error
	Swap(ctx context.Context, userID, fromID, toID int) (model.Course, error)
	UpdateQuota(ctx context.Context, courseID int, req model.UpdateQuotaRequest) (model.Course, error)
}

// UserRepository adalah kontrak repository tabel users.
type UserRepository interface {
	Register(ctx context.Context, req model.RegisterRequest, passwordHash string) (int, error)
	GetByUsername(ctx context.Context, username string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	List(ctx context.Context, limit int) ([]model.User, error)
	FindAfterCursor(ctx context.Context, cursor time.Time, cursorID, limit int) ([]model.User, error)
}

// Compile-time assertion: implementasi PostgreSQL wajib memenuhi kontrak.
var (
	_ CourseRepository = (*PostgresCourseRepository)(nil)
	_ UserRepository   = (*PostgresUserRepository)(nil)
)
