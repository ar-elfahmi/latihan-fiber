//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
)

// newCourseRepo membuat repository kelas dengan aturan batas SKS yang sama
// dengan production (diinjeksi dari app/service).
func newCourseRepo(t *testing.T) repository.CourseRepository {
	t.Helper()

	repo, err := repository.NewPostgresCourseRepository(pool, repository.CreditRules{
		WithinCreditLimit: service.WithinCreditLimit,
		CreditsAfterSwap:  service.CreditsAfterSwap,
	})
	if err != nil {
		t.Fatalf("membuat repository kelas: %v", err)
	}
	return repo
}

// TestIncrementEnrolled_StopsAtQuota memastikan penambahan jumlah peserta
// BERHENTI di kuota: kelas dengan kuota 2 hanya menerima 2 peserta, peserta
// ketiga ditolak dan enrolled_count tidak naik.
func TestIncrementEnrolled_StopsAtQuota(t *testing.T) {
	ctx := context.Background()
	course := seedCourse(t, 3, 2)
	repo := newCourseRepo(t)

	students := []model.User{seedStudent(t), seedStudent(t), seedStudent(t)}
	for i := 0; i < 2; i++ {
		if _, err := repo.Enroll(ctx, students[i].ID, course.ID); err != nil {
			t.Fatalf("kursi ke-%d: error tidak diharapkan: %v", i+1, err)
		}
	}

	_, err := repo.Enroll(ctx, students[2].ID, course.ID)
	if !errors.Is(err, repository.ErrCourseFull) {
		t.Fatalf("kursi ke-3: error = %v, mau ErrCourseFull", err)
	}

	got, err := repo.FindByID(ctx, course.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.EnrolledCount != 2 {
		t.Errorf("enrolled_count = %d, mau 2 (penambahan berhenti di kuota)", got.EnrolledCount)
	}
}

// TestInsertEnrollment_Duplicate memastikan enrollment ganda untuk pasangan
// (kelas, mahasiswa) yang sama ditolak.
func TestInsertEnrollment_Duplicate(t *testing.T) {
	ctx := context.Background()
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)
	repo := newCourseRepo(t)

	if _, err := repo.Enroll(ctx, student.ID, course.ID); err != nil {
		t.Fatalf("pendaftaran pertama: %v", err)
	}

	_, err := repo.Enroll(ctx, student.ID, course.ID)
	if !errors.Is(err, repository.ErrAlreadyEnrolled) {
		t.Fatalf("pendaftaran kedua: error = %v, mau ErrAlreadyEnrolled", err)
	}

	if n := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, course.ID); n != 1 {
		t.Errorf("baris enrollment = %d, mau 1 (duplikat tidak boleh masuk)", n)
	}
}

// TestWithTx_RollsBackEveryStep memastikan transaksi yang gagal membatalkan
// SEMUA langkahnya, termasuk catatan di outbox. Inilah dasar pola outbox.
func TestWithTx_RollsBackEveryStep(t *testing.T) {
	ctx := context.Background()
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)
	txm := repository.NewTxManager(pool)

	payload, err := json.Marshal(model.EnrollmentEvent{UserID: student.ID, CourseID: course.ID})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	errStop := errors.New("gagal yang disengaja")
	err = txm.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO enrollments (user_id, course_id) VALUES ($1, $2)`,
			student.ID, course.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE courses SET enrolled_count = enrolled_count + 1 WHERE id = $1`,
			course.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
			model.EventEnrollmentCreated, payload); err != nil {
			return err
		}
		return errStop
	})
	if !errors.Is(err, errStop) {
		t.Fatalf("WithTx: error = %v, mau errStop", err)
	}

	if n := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1`, student.ID); n != 0 {
		t.Errorf("baris enrollment = %d, mau 0 setelah rollback", n)
	}
	if n := countRows(t, `SELECT COUNT(*) FROM outbox WHERE payload->>'user_id' = $1`,
		strconv.Itoa(student.ID)); n != 0 {
		t.Errorf("baris outbox = %d, mau 0 setelah rollback", n)
	}
	var enrolledCount int
	if err := pool.QueryRow(ctx, `SELECT enrolled_count FROM courses WHERE id = $1`,
		course.ID).Scan(&enrolledCount); err != nil {
		t.Fatalf("baca enrolled_count: %v", err)
	}
	if enrolledCount != 0 {
		t.Errorf("enrolled_count = %d, mau 0 setelah rollback", enrolledCount)
	}
}

// TestClaim_OnlyOnce ensures satu event outbox hanya bisa diklaim sekali
// (ON CONFLICT DO NOTHING di tabel notifications).
func TestClaim_OnlyOnce(t *testing.T) {
	ctx := context.Background()
	student := seedStudent(t)

	payload, err := json.Marshal(model.EnrollmentEvent{UserID: student.ID, CourseID: 1})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	var outboxID int64
	if err := pool.QueryRow(ctx,
		`INSERT INTO outbox (event_type, payload) VALUES ($1, $2) RETURNING id`,
		model.EventEnrollmentCreated, payload).Scan(&outboxID); err != nil {
		t.Fatalf("seed outbox: %v", err)
	}

	repo := repository.PostgresNotificationRepository{DB: pool}
	n := model.Notification{OutboxID: outboxID, UserID: student.ID, Channel: "email",
		Subject: "s", Body: "b"}

	first, err := repo.Claim(ctx, n)
	if err != nil || !first {
		t.Fatalf("claim pertama = %v, %v; mau true, nil", first, err)
	}
	second, err := repo.Claim(ctx, n)
	if err != nil || second {
		t.Fatalf("claim kedua = %v, %v; mau false, nil", second, err)
	}
	if rows := countRows(t, `SELECT COUNT(*) FROM notifications WHERE outbox_id = $1`,
		outboxID); rows != 1 {
		t.Errorf("baris notifications = %d, mau 1", rows)
	}
}

// TestEnrollEndpoint_WritesOutboxOnlyOnSuccess memastikan endpoint enroll
// menulis event outbox HANYA kalau pendaftaran benar-benar berhasil.
func TestEnrollEndpoint_WritesOutboxOnlyOnSuccess(t *testing.T) {
	app := newAPI(t)
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)
	path := "/courses/" + strconv.Itoa(course.ID) + "/enroll"

	status, body := request(t, app, "POST", path, student.ID, "")
	if status != fiber.StatusCreated {
		t.Fatalf("pendaftaran pertama: status = %d, mau 201 (body: %v)", status, body)
	}

	status, body = request(t, app, "POST", path, student.ID, "")
	if status != fiber.StatusConflict || body["code"] != "ALREADY_ENROLLED" {
		t.Fatalf("pendaftaran kedua: dapat %d %v, mau 409 ALREADY_ENROLLED", status, body["code"])
	}

	events := countRows(t,
		`SELECT COUNT(*) FROM outbox WHERE event_type = $1 AND payload->>'user_id' = $2`,
		model.EventEnrollmentCreated, strconv.Itoa(student.ID))
	if events != 1 {
		t.Errorf("event enrollment.created = %d, mau 1 (pendaftaran gagal tidak boleh meninggalkan jejak)", events)
	}
}

// TestEnrollEndpoint_ConcurrentCreditLimit memastikan N pendaftaran yang
// berjalan SERENTAK tidak ada yang melewati batas 24 SKS.
//
// Satu mahasiswa mengirim sepuluh pendaftaran serentak ke sepuluh kelas 3 SKS.
// Batasnya 24 SKS, maka tepat delapan yang boleh berhasil.
//
// Test diulang dua ronde: peluang tabrakan harus muncul di salah satu ronde,
// sehingga mutasi "hapus FOR UPDATE" tidak bisa lolos tanpa terdeteksi.
func TestEnrollEndpoint_ConcurrentCreditLimit(t *testing.T) {
	app := newAPI(t)

	// Hangatkan pool lebih dulu: tanpa ini koneksi baru dibuat saat test
	// berjalan, sehingga request tidak benar-benar bertabrakan.
	warm := sync.WaitGroup{}
	for i := 0; i < 12; i++ {
		warm.Add(1)
		go func() {
			defer warm.Done()
			_, _ = pool.Exec(context.Background(), "SELECT 1")
		}()
	}
	warm.Wait()

	for ronde := 1; ronde <= 2; ronde++ {
		student := seedStudent(t)

		const attempts = 10
		courseIDs := make([]int, 0, attempts)
		for i := 0; i < attempts; i++ {
			courseIDs = append(courseIDs, seedCourse(t, 3, 40).ID)
		}

		var (
			wg      sync.WaitGroup
			mu      sync.Mutex
			results = map[int]int{}
			start   = make(chan struct{})
		)

		for _, courseID := range courseIDs {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				<-start // aba-aba: semua request berangkat bersamaan
				status, _ := request(t, app, "POST", "/courses/"+strconv.Itoa(id)+"/enroll", student.ID, "")
				mu.Lock()
				results[status]++
				mu.Unlock()
			}(courseID)
		}
		close(start)
		wg.Wait()

		maxCourses := service.MaxCreditsPerSemester / 3
		if results[fiber.StatusCreated] != maxCourses {
			t.Errorf("ronde %d: berhasil = %d, mau %d (hasil lengkap: %v)",
				ronde, results[fiber.StatusCreated], maxCourses, results)
		}
		if enrolled := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1`,
			student.ID); enrolled != maxCourses {
			t.Errorf("ronde %d: baris enrollments = %d, mau %d", ronde, enrolled, maxCourses)
		}

		var totalCredits int
		if err := pool.QueryRow(context.Background(),
			`SELECT COALESCE(SUM(c.credits), 0) FROM enrollments e
			 JOIN courses c ON c.id = e.course_id WHERE e.user_id = $1`,
			student.ID).Scan(&totalCredits); err != nil {
			t.Fatalf("hitung total SKS: %v", err)
		}
		if totalCredits > service.MaxCreditsPerSemester {
			t.Errorf("ronde %d: total SKS = %d, harus <= %d (batas SKS terlampaui)",
				ronde, totalCredits, service.MaxCreditsPerSemester)
		}
	}
}
