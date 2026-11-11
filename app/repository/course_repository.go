package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
)

var (
	ErrCourseFull      = errors.New("kuota kelas penuh")
	ErrAlreadyEnrolled = errors.New("sudah terdaftar di kelas ini")
	ErrNotEnrolled     = errors.New("belum terdaftar di kelas ini")
	ErrCreditLimit     = errors.New("batas 24 SKS terlampaui")
	ErrQuotaConflict   = errors.New("versi kelas berubah atau kuota di bawah jumlah peserta")
)

const courseColumns = "id, code, name, credits, quota, enrolled_count, version, created_at"

type PostgresCourseRepository struct {
	DB *pgxpool.Pool
}

func scanCourse(row pgx.Row) (model.Course, error) {
	var course model.Course
	err := row.Scan(&course.ID, &course.Code, &course.Name, &course.Credits,
		&course.Quota, &course.EnrolledCount, &course.Version, &course.CreatedAt)
	return course, err
}

func (r PostgresCourseRepository) List(ctx context.Context) ([]model.Course, error) {
	rows, err := r.DB.Query(ctx, "SELECT "+courseColumns+" FROM courses ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := []model.Course{}
	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, rows.Err()
}

func (r PostgresCourseRepository) Enroll(ctx context.Context, userID, courseID int) (model.Course, error) {
	var course model.Course
	outbox := OutboxRepository{DB: r.DB}
	err := pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		if err := lockKRS(ctx, tx, userID, courseID, courseID); err != nil {
			return err
		}
		class, err := scanCourse(tx.QueryRow(ctx, "SELECT "+courseColumns+" FROM courses WHERE id = $1", courseID))
		if err != nil {
			return err
		}
		if err := outbox.WithTx(tx).Insert(ctx, model.EventEnrollmentCreated, model.EnrollmentEvent{
			UserID: userID, CourseID: class.ID, CourseCode: class.Code, CourseName: class.Name,
		}); err != nil {
			return err
		}
		course, err = enroll(ctx, tx, userID, courseID)
		return err
	})
	return course, err
}

func (r PostgresCourseRepository) Drop(ctx context.Context, userID, courseID int) error {
	return pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		if err := lockKRS(ctx, tx, userID, courseID, courseID); err != nil {
			return err
		}
		if err := drop(ctx, tx, userID, courseID); err != nil {
			return err
		}
		class, err := scanCourse(tx.QueryRow(ctx, "SELECT "+courseColumns+" FROM courses WHERE id = $1", courseID))
		if err != nil {
			return err
		}
		return (OutboxRepository{DB: r.DB}).WithTx(tx).Insert(ctx, model.EventEnrollmentDropped, model.EnrollmentEvent{
			UserID: userID, CourseID: class.ID, CourseCode: class.Code, CourseName: class.Name,
		})
	})
}

func (r PostgresCourseRepository) Swap(ctx context.Context, userID, fromID, toID int) (model.Course, error) {
	var course model.Course
	err := pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		if err := lockKRS(ctx, tx, userID, fromID, toID); err != nil {
			return err
		}
		if err := drop(ctx, tx, userID, fromID); err != nil {
			return err
		}
		from, err := scanCourse(tx.QueryRow(ctx, "SELECT "+courseColumns+" FROM courses WHERE id = $1", fromID))
		if err != nil {
			return err
		}
		course, err = enroll(ctx, tx, userID, toID)
		if err != nil {
			return err
		}
		outbox := OutboxRepository{DB: r.DB}
		event := model.EnrollmentEvent{
			UserID: userID, CourseID: course.ID, CourseCode: course.Code, CourseName: course.Name,
			FromCourseCode: from.Code, FromCourseName: from.Name,
		}
		return outbox.WithTx(tx).Insert(ctx, model.EventEnrollmentSwapped, event)
	})
	return course, err
}

// Kunci akun sebelum kelas agar dua perubahan KRS milik akun yang sama
// tidak bisa sama-sama lolos batas SKS. Urutkan kunci kelas untuk swap.
func lockKRS(ctx context.Context, tx pgx.Tx, userID, firstID, secondID int) error {
	var id int
	if err := tx.QueryRow(ctx, "SELECT id FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&id); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, "SELECT id FROM courses WHERE id = $1 FOR UPDATE", min(firstID, secondID)).Scan(&id); err != nil {
		return err
	}
	if firstID != secondID {
		return tx.QueryRow(ctx, "SELECT id FROM courses WHERE id = $1 FOR UPDATE", max(firstID, secondID)).Scan(&id)
	}
	return nil
}

func enroll(ctx context.Context, tx pgx.Tx, userID, courseID int) (model.Course, error) {
	inserted, err := tx.Exec(ctx, `INSERT INTO enrollments (user_id, course_id)
		VALUES ($1, $2) ON CONFLICT (course_id, user_id) DO NOTHING`, userID, courseID)
	if err != nil {
		return model.Course{}, err
	}
	if inserted.RowsAffected() == 0 {
		return model.Course{}, ErrAlreadyEnrolled
	}

	var credits int
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(c.credits), 0) FROM enrollments e
		JOIN courses c ON c.id = e.course_id WHERE e.user_id = $1`, userID).Scan(&credits)
	if err != nil {
		return model.Course{}, err
	}
	if credits > 24 {
		return model.Course{}, ErrCreditLimit
	}
	course, err := scanCourse(tx.QueryRow(ctx, `UPDATE courses SET enrolled_count = enrolled_count + 1
		WHERE id = $1 AND enrolled_count < quota RETURNING `+courseColumns, courseID))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Course{}, ErrCourseFull
	}
	return course, err
}

func drop(ctx context.Context, tx pgx.Tx, userID, courseID int) error {
	deleted, err := tx.Exec(ctx, "DELETE FROM enrollments WHERE user_id = $1 AND course_id = $2", userID, courseID)
	if err != nil {
		return err
	}
	if deleted.RowsAffected() == 0 {
		return ErrNotEnrolled
	}
	_, err = tx.Exec(ctx, "UPDATE courses SET enrolled_count = enrolled_count - 1 WHERE id = $1", courseID)
	return err
}

func (r PostgresCourseRepository) UpdateQuota(ctx context.Context, courseID int, req model.UpdateQuotaRequest) (model.Course, error) {
	var course model.Course
	err := pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		var err error
		course, err = scanCourse(tx.QueryRow(ctx, `UPDATE courses SET quota = $2, version = version + 1
			WHERE id = $1 AND version = $3 AND enrolled_count <= $2 RETURNING `+courseColumns,
			courseID, req.Quota, req.Version))
		if errors.Is(err, pgx.ErrNoRows) {
			var id int
			if err := tx.QueryRow(ctx, "SELECT id FROM courses WHERE id = $1", courseID).Scan(&id); err != nil {
				return err
			}
			return ErrQuotaConflict
		}
		if err != nil {
			return err
		}
		var recipients []int
		err = tx.QueryRow(ctx, `SELECT array_agg(user_id ORDER BY user_id)
			FROM enrollments WHERE course_id = $1`, courseID).Scan(&recipients)
		if err != nil {
			return err
		}
		return (OutboxRepository{DB: tx}).Insert(ctx, model.EventCourseQuotaChanged, model.EnrollmentEvent{
			CourseID: course.ID, CourseCode: course.Code, CourseName: course.Name,
			Quota: course.Quota, RecipientIDs: recipients,
		})
	})
	return course, err
}
