package service

import (
	"reflect"
	"testing"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
)

// TestMaxCreditsPerSemester menjaga angka batas SKS tetap satu sumber
// kebenaran: 24. Angka ini juga ditulis di domain.MaxCreditsPerSemester.
func TestMaxCreditsPerSemester(t *testing.T) {
	t.Parallel()

	if MaxCreditsPerSemester != 24 {
		t.Fatalf("MaxCreditsPerSemester = %d, mau 24", MaxCreditsPerSemester)
	}
}

// TestWithinCreditLimit menguji batas SKS tepat di batas: 24 SKS lolos,
// 25 SKS ditolak (temuan B1 #8 — tanpa ini mutasi "<=" menjadi "<" lolos).
func TestWithinCreditLimit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		current int
		delta   int
		want    bool
	}{
		{name: "nol SKS + 24 SKS lolos", current: 0, delta: 24, want: true},
		{name: "tepat di batas: 24 + 0 lolos", current: 24, delta: 0, want: true},
		{name: "tepat di batas: 21 + 3 lolos", current: 21, delta: 3, want: true},
		{name: "satu di atas batas ditolak", current: 24, delta: 1, want: false},
		{name: "melebihi batas ditolak", current: 24, delta: 3, want: false},
		{name: "nol SKS + 25 SKS ditolak", current: 0, delta: 25, want: false},
		{name: "kelas dilepas (delta negatif) selalu lolos", current: 24, delta: -3, want: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := WithinCreditLimit(tc.current, tc.delta)
			if got != tc.want {
				t.Fatalf("WithinCreditLimit(%d, %d) = %v, mau %v", tc.current, tc.delta, got, tc.want)
			}
		})
	}
}

// TestCreditsAfterSwap menguji perhitungan SKS setelah tukar kelas.
func TestCreditsAfterSwap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		current   int
		oldCourse model.Course
		newCourse model.Course
		want      int
	}{
		{
			name:      "tukar naik: 18 SKS, kelas asal 3 SKS, kelas tujuan 4 SKS",
			current:   18,
			oldCourse: model.Course{Credits: 3},
			newCourse: model.Course{Credits: 4},
			want:      19,
		},
		{
			name:      "tukar turun: 18 SKS, kelas asal 4 SKS, kelas tujuan 2 SKS",
			current:   18,
			oldCourse: model.Course{Credits: 4},
			newCourse: model.Course{Credits: 2},
			want:      16,
		},
		{
			name:      "tukar dengan SKS sama: 18 SKS, 3 SKS ke 3 SKS",
			current:   18,
			oldCourse: model.Course{Credits: 3},
			newCourse: model.Course{Credits: 3},
			want:      18,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CreditsAfterSwap(tc.current, tc.oldCourse, tc.newCourse)
			if got != tc.want {
				t.Fatalf("CreditsAfterSwap(%d, %d SKS, %d SKS) = %d, mau %d",
					tc.current, tc.oldCourse.Credits, tc.newCourse.Credits, got, tc.want)
			}
		})
	}
}

// TestNormalizeTags menguji normalisasi tag: trim, huruf kecil, buang tag
// kosong, hapus duplikat, dan urutan stabil.
func TestNormalizeTags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tags []string
		want []string
	}{
		{
			name: "duplikat dihapus",
			tags: []string{"go", "go", "sql"},
			want: []string{"go", "sql"},
		},
		{
			name: "spasi di kedua ujung dipangkas",
			tags: []string{"  go  ", " sql "},
			want: []string{"go", "sql"},
		},
		{
			name: "huruf besar menjadi kecil",
			tags: []string{"Go", "SQL", "PostgreSQL"},
			want: []string{"go", "sql", "postgresql"},
		},
		{
			name: "tag kosong dibuang",
			tags: []string{"go", "", "   ", "sql"},
			want: []string{"go", "sql"},
		},
		{
			name: "urutan kemunculan pertama dipertahankan",
			tags: []string{"SQL", "go", "sql", "Go", "mongo"},
			want: []string{"sql", "go", "mongo"},
		},
		{
			name: "daftar kosong menghasilkan daftar kosong",
			tags: []string{},
			want: []string{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assertTags(t, NormalizeTags(tc.tags), tc.want)
		})
	}
}

// TestCreditLimitErrorMessageBelumBerubah menjaga pesan error batas SKS
// tetap sama seperti kontrak endpoint (409 + "batas 24 SKS terlampaui"),
// walaupun pesannya kini disusun dari konstanta tunggal.
func TestCreditLimitErrorMessageBelumBerubah(t *testing.T) {
	t.Parallel()

	if got := repository.ErrCreditLimit.Error(); got != "batas 24 SKS terlampaui" {
		t.Fatalf("pesan error batas SKS = %q, mau %q", got, "batas 24 SKS terlampaui")
	}
}

// assertTags membandingkan dua daftar tag.
func assertTags(t *testing.T, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeTags = %v, mau %v", got, want)
	}
}
