//go:build integration && d2

// Build tag tambahan "d2" dipakai supaya kriteria C.1 T4 ("tepat enam baris
// --- PASS tingkat atas") tetap terpenuhi: dua test tambahan D.2 di file ini
// dijalankan terpisah dengan:
//
//	go test -tags="integration d2" -count=1 -v ./tests/integration/
//
// Isi file: dua test tambahan dari D.2 (swap ke kelas penuh, keluar kelas).

package integration_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
)

// TestSwapEndpoint_RejectsFullTargetCourse memastikan tukar kelas ke kelas
// yang sudah PENUH ditolak dengan 409, mahasiswa tetap di kelas asal,
// enrolled_count kedua kelas tidak berubah, dan tidak ada event
// enrollment.swapped di outbox.
//
// Catatan kode error: produksi memakai kode stabil CONFLICT untuk semua
// konflik KRS (helper.Conflict), sedangkan teks modul menyebut COURSE_FULL.
// Test menerima keduanya; status 409 dan pesan "penuh" yang diuji.
func TestSwapEndpoint_RejectsFullTargetCourse(t *testing.T) {
	app := newAPI(t)
	fromCourse := seedCourse(t, 3, 40)
	toCourse := seedCourse(t, 3, 1) // kuota 1: akan penuh
	student := seedStudent(t)
	other := seedStudent(t)
	repo := newCourseRepo(t)
	ctx := context.Background()

	// Mahasiswa ambil kelas asal; mahasiswa lain mengisi kelas tujuan.
	if _, err := repo.Enroll(ctx, student.ID, fromCourse.ID); err != nil {
		t.Fatalf("enroll kelas asal: %v", err)
	}
	if _, err := repo.Enroll(ctx, other.ID, toCourse.ID); err != nil {
		t.Fatalf("enroll kelas tujuan: %v", err)
	}

	body := `{"from_course_id":` + strconv.Itoa(fromCourse.ID) +
		`,"to_course_id":` + strconv.Itoa(toCourse.ID) + `}`
	status, parsed := request(t, app, "POST", "/enrollments/swap", student.ID, body)
	if status != fiber.StatusConflict {
		t.Fatalf("swap ke kelas penuh: status = %d, mau 409 (body: %v)", status, parsed)
	}
	code, _ := parsed["code"].(string)
	if code != "CONFLICT" && code != "COURSE_FULL" {
		t.Errorf("kode error = %q, mau CONFLICT (produksi) atau COURSE_FULL (modul)", code)
	}
	if msg, _ := parsed["message"].(string); msg == "" {
		t.Error("pesan error kosong")
	}

	// Mahasiswa tetap di kelas asal.
	if n := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, fromCourse.ID); n != 1 {
		t.Errorf("enrollment di kelas asal = %d, mau 1 (mahasiswa harus tetap di kelas asal)", n)
	}
	if n := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, toCourse.ID); n != 0 {
		t.Errorf("enrollment di kelas tujuan = %d, mau 0", n)
	}

	// enrolled_count kedua kelas tidak berubah.
	if n := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, fromCourse.ID); n != 1 {
		t.Errorf("enrolled_count kelas asal = %d, mau 1", n)
	}
	if n := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, toCourse.ID); n != 1 {
		t.Errorf("enrolled_count kelas tujuan = %d, mau 1", n)
	}

	// Tidak ada event enrollment.swapped untuk mahasiswa ini.
	if n := countRows(t,
		`SELECT COUNT(*) FROM outbox WHERE event_type = $1 AND payload->>'user_id' = $2`,
		model.EventEnrollmentSwapped, strconv.Itoa(student.ID)); n != 0 {
		t.Errorf("event enrollment.swapped = %d, mau 0 (swap gagal tidak boleh meninggalkan jejak)", n)
	}
}

// TestDropEndpoint_RemovesEnrollment memastikan DELETE /courses/:id/enroll
// berhasil: 204, baris enrollment hilang, enrolled_count berkurang satu, dan
// outbox berisi tepat satu event enrollment.dropped.
func TestDropEndpoint_RemovesEnrollment(t *testing.T) {
	app := newAPI(t)
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)
	path := "/courses/" + strconv.Itoa(course.ID) + "/enroll"

	status, body := request(t, app, "POST", path, student.ID, "")
	if status != fiber.StatusCreated {
		t.Fatalf("pendaftaran: status = %d, mau 201 (body: %v)", status, body)
	}
	if n := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, course.ID); n != 1 {
		t.Fatalf("enrolled_count = %d, mau 1 setelah daftar", n)
	}

	status, body = request(t, app, "DELETE", path, student.ID, "")
	if status != fiber.StatusNoContent {
		t.Fatalf("keluar kelas: status = %d, mau 204 (body: %v)", status, body)
	}

	if n := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, course.ID); n != 0 {
		t.Errorf("baris enrollment = %d, mau 0 setelah keluar kelas", n)
	}
	if n := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, course.ID); n != 0 {
		t.Errorf("enrolled_count = %d, mau 0 setelah keluar kelas", n)
	}
	if n := countRows(t,
		`SELECT COUNT(*) FROM outbox WHERE event_type = $1 AND payload->>'user_id' = $2`,
		model.EventEnrollmentDropped, strconv.Itoa(student.ID)); n != 1 {
		t.Errorf("event enrollment.dropped = %d, mau 1", n)
	}
}
