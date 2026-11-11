package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/domain"
	"latihan-fiber/app/model"
)

var (
	ErrCourseFull      = errors.New("kuota kelas penuh")
	ErrAlreadyEnrolled = errors.New("sudah terdaftar di kelas ini")
	ErrNotEnrolled     = errors.New("belum terdaftar di kelas ini")
	// Pesan batas SKS disusun dari konstanta tunggal supaya angka 24 tidak
	// ditulis dua kali; teksnya tetap "batas 24 SKS terlampaui".
	ErrCreditLimit   = errors.New(fmt.Sprintf("batas %d SKS terlampaui", domain.MaxCreditsPerSemester))
	ErrQuotaConflict = errors.New("versi kelas berubah atau kuota di bawah jumlah peserta")
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
	outbox := PostgresOutboxRepository{DB: r.DB}
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
		if err := insertEnrollment(ctx, tx, userID, courseID); err != nil {
			return err
		}
		// Batas SKS diperiksa di dalam transaksi: total SKS setelah diambil
		// kelas ini harus tetap di dalam batas.
		total, err := totalCredits(ctx, tx, userID)
		if err != nil {
			return err
		}
		if !domain.WithinCreditLimit(total-class.Credits, class.Credits) {
			return ErrCreditLimit
		}
		course, err = bumpEnrolledCount(ctx, tx, courseID)
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
		return (PostgresOutboxRepository{DB: r.DB}).WithTx(tx).Insert(ctx, model.EventEnrollmentDropped, model.EnrollmentEvent{
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
		// Total SKS sebelum tukar; dipakai menghitung SKS setelah tukar.
		totalBefore, err := totalCredits(ctx, tx, userID)
		if err != nil {
			return err
		}
		if err := drop(ctx, tx, userID, fromID); err != nil {
			return err
		}
		from, err := scanCourse(tx.QueryRow(ctx, "SELECT "+courseColumns+" FROM courses WHERE id = $1", fromID))
		if err != nil {
			return err
		}
		// Kelas tujuan dibaca lebih dulu untuk menghitung SKS setelah tukar.
		// Kalau kelas tujuan tidak ada, insert di bawah yang menolak (perilaku
		// lama: error foreign key) sehingga respons tidak berubah.
		to, toFound := model.Course{}, false
		scanned, err := scanCourse(tx.QueryRow(ctx, "SELECT "+courseColumns+" FROM courses WHERE id = $1", toID))
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		} else {
			to, toFound = scanned, true
		}
		if err := insertEnrollment(ctx, tx, userID, toID); err != nil {
			return err
		}
		if toFound {
			// SKS setelah tukar = SKS lama - SKS kelas asal + SKS kelas tujuan.
			afterSwap := domain.CreditsAfterSwap(totalBefore, from, to)
			if afterSwap > domain.MaxCreditsPerSemester {
				return ErrCreditLimit
			}
		}
		course, err = bumpEnrolledCount(ctx, tx, toID)
		if err != nil {
			return err
		}
		outbox := PostgresOutboxRepository{DB: r.DB}
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

// insertEnrollment menulis baris enrollment mahasiswa pada sebuah kelas.
// ErrAlreadyEnrolled kalau mahasiswa sudah terdaftar di kelas itu.
func insertEnrollment(ctx context.Context, tx pgx.Tx, userID, courseID int) error {
	inserted, err := tx.Exec(ctx, `INSERT INTO enrollments (user_id, course_id)
		VALUES ($1, $2) ON CONFLICT (course_id, user_id) DO NOTHING`, userID, courseID)
	if err != nil {
		return err
	}
	if inserted.RowsAffected() == 0 {
		return ErrAlreadyEnrolled
	}
	return nil
}

// totalCredits menjumlahkan SKS seluruh kelas yang sedang diambil mahasiswa.
func totalCredits(ctx context.Context, tx pgx.Tx, userID int) (int, error) {
	var total int
	err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(c.credits), 0) FROM enrollments e
		JOIN courses c ON c.id = e.course_id WHERE e.user_id = $1`, userID).Scan(&total)
	return total, err
}

// bumpEnrolledCount menambah jumlah peserta kelas. ErrCourseFull kalau
// kuota sudah habis.
func bumpEnrolledCount(ctx context.Context, tx pgx.Tx, courseID int) (model.Course, error) {
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
		return (PostgresOutboxRepository{DB: tx}).Insert(ctx, model.EventCourseQuotaChanged, model.EnrollmentEvent{
			CourseID: course.ID, CourseCode: course.Code, CourseName: course.Name,
			Quota: course.Quota, RecipientIDs: recipients,
		})
	})
	return course, err
}
