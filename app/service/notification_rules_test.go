package service

import (
	"errors"
	"strings"
	"testing"

	"latihan-fiber/app/model"
)

// TestBuildMessage menguji BuildMessage untuk keempat jenis event KRS.
// File ini menjaga coverage BuildMessage = 100% (kriteria C.1 T3).
func TestBuildMessage(t *testing.T) {
	t.Parallel()

	user := model.User{ID: 9, Username: "budi", Email: "budi@example.com"}
	event := model.EnrollmentEvent{
		UserID:         9,
		CourseID:       1,
		CourseCode:     "IF101",
		CourseName:     "Pemrograman Backend Lanjut",
		FromCourseCode: "IF102",
		FromCourseName: "Basis Data Lanjut",
		Quota:          30,
	}

	cases := []struct {
		name        string
		eventType   string
		wantSubject string
		wantBody    string
	}{
		{
			name:        "pendaftaran kelas",
			eventType:   model.EventEnrollmentCreated,
			wantSubject: "KRS: berhasil terdaftar di IF101 Pemrograman Backend Lanjut",
			wantBody:    "Halo budi, Anda tercatat sebagai peserta kelas IF101 Pemrograman Backend Lanjut.",
		},
		{
			name:        "keluar kelas",
			eventType:   model.EventEnrollmentDropped,
			wantSubject: "KRS: keluar dari IF101 Pemrograman Backend Lanjut",
			wantBody:    "Halo budi, Anda sudah tidak terdaftar di kelas IF101 Pemrograman Backend Lanjut.",
		},
		{
			name:        "pindah kelas",
			eventType:   model.EventEnrollmentSwapped,
			wantSubject: "KRS: pindah dari IF102 Basis Data Lanjut ke IF101 Pemrograman Backend Lanjut",
			wantBody:    "Halo budi, Anda telah pindah dari kelas IF102 Basis Data Lanjut ke kelas IF101 Pemrograman Backend Lanjut.",
		},
		{
			name:        "kuota kelas berubah",
			eventType:   model.EventCourseQuotaChanged,
			wantSubject: "KRS: kuota IF101 Pemrograman Backend Lanjut berubah menjadi 30",
			wantBody:    "Halo budi, kuota kelas IF101 Pemrograman Backend Lanjut sekarang 30.",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			msg, err := BuildMessage(tc.eventType, event, user)
			if err != nil {
				t.Fatalf("BuildMessage(%q): error tidak diharapkan: %v", tc.eventType, err)
			}
			if msg.To != user.Email {
				t.Errorf("To = %q, mau %q", msg.To, user.Email)
			}
			if msg.Subject != tc.wantSubject {
				t.Errorf("Subject = %q, mau %q", msg.Subject, tc.wantSubject)
			}
			if msg.Body != tc.wantBody {
				t.Errorf("Body = %q, mau %q", msg.Body, tc.wantBody)
			}
		})
	}
}

// TestBuildMessageUnknownEvent memastikan event yang tidak dikenal ditolak
// dengan error yang bisa dikenali lewat errors.Is.
func TestBuildMessageUnknownEvent(t *testing.T) {
	t.Parallel()

	_, err := BuildMessage("event.tidak.dikenal", model.EnrollmentEvent{}, model.User{})
	if err == nil {
		t.Fatal("BuildMessage untuk event asing harus error")
	}
	if !errors.Is(err, errUnknownEvent) {
		t.Fatalf("error = %v, mau errUnknownEvent", err)
	}
	if !strings.Contains(err.Error(), "event.tidak.dikenal") {
		t.Errorf("pesan error = %q, mau memuat nama event", err.Error())
	}
}
