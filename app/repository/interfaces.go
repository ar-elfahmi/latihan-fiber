package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

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

// StudentRepository adalah kontrak repository tabel students.
type StudentRepository interface {
	List(ctx context.Context, limit int) ([]model.Student, error)
	FindAfterCursor(ctx context.Context, cursorID, limit int) ([]model.Student, error)
	GetByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Replace(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

// PrestasiRepository adalah kontrak repository tabel prestasi.
type PrestasiRepository interface {
	GetAll(ctx context.Context, q model.ListQuery) ([]model.Prestasi, int, error)
	GetByStudentID(ctx context.Context, studentID int) ([]model.Prestasi, error)
	Create(ctx context.Context, p model.Prestasi) (model.Prestasi, error)
}

// NotificationRepository adalah kontrak repository tabel notifications.
// WithTx mengembalikan repo yang terikat ke transaksi berjalan.
type NotificationRepository interface {
	WithTx(tx pgx.Tx) NotificationRepository
	Claim(ctx context.Context, n model.Notification) (bool, error)
	FindByUser(ctx context.Context, userID, limit int) ([]model.Notification, error)
	Count(ctx context.Context) (int, error)
}

// OutboxRepository adalah kontrak repository tabel outbox.
// WithTx mengembalikan repo yang terikat ke transaksi berjalan.
type OutboxRepository interface {
	WithTx(tx pgx.Tx) OutboxRepository
	Insert(ctx context.Context, eventType string, event model.EnrollmentEvent) error
	FetchUnpublished(ctx context.Context, afterID int64, limit int) ([]model.OutboxEvent, error)
	MarkPublished(ctx context.Context, ids []int64) error
	Stats(ctx context.Context) (int, int, time.Duration, error)
}

// RBACRepository adalah kontrak repository pemetaan role -> permission.
type RBACRepository interface {
	GetPermissionsByUserID(ctx context.Context, userID int) (map[string]bool, error)
	GetRolesByUserID(ctx context.Context, userID int) ([]string, error)
}

// Compile-time assertion: implementasi PostgreSQL wajib memenuhi kontrak.
var (
	_ CourseRepository       = (*PostgresCourseRepository)(nil)
	_ UserRepository         = (*PostgresUserRepository)(nil)
	_ StudentRepository      = (*PostgresStudentRepository)(nil)
	_ PrestasiRepository     = (*PostgresPrestasiRepository)(nil)
	_ NotificationRepository = (*PostgresNotificationRepository)(nil)
	_ OutboxRepository       = (*PostgresOutboxRepository)(nil)
	_ RBACRepository         = (*PostgresRBACRepository)(nil)
)
