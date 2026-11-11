package service

import (
	"strings"
	"testing"

	"latihan-fiber/app/model"
	"latihan-fiber/helper"
)

// Aturan ulasan kelas (S-004) diuji di sini pada level aturan:
//   - rating harus 1..5 (tag validate "required,min=1,max=5"),
//   - komentar 10..1000 karakter,
//   - maksimal 10 tag, tiap tag 2..50 karakter,
//   - tag disimpan dalam bentuk ternormalisasi (NormalizeTags).
//
// Semua aturan dibaca dari tag validate pada model.CreateReviewRequest
// sehingga test ini mengunci satu sumber kebenaran.

func TestReviewRatingRules(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		rating    int
		wantValid bool
	}{
		{name: "rating minimum 1", rating: 1, wantValid: true},
		{name: "rating tengah 3", rating: 3, wantValid: true},
		{name: "rating maksimum 5", rating: 5, wantValid: true},
		{name: "rating 0 ditolak", rating: 0, wantValid: false},
		{name: "rating 6 ditolak", rating: 6, wantValid: false},
		{name: "rating negatif ditolak", rating: -1, wantValid: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fields := helper.ValidateStruct(model.CreateReviewRequest{
				Rating:  tc.rating,
				Comment: "Komentar yang cukup panjang untuk lolos.",
			})
			got := len(fields) == 0
			if got != tc.wantValid {
				t.Fatalf("rating %d valid = %v, mau %v (fields: %v)", tc.rating, got, tc.wantValid, fields)
			}
		})
	}
}

func TestReviewCommentRules(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		comment   string
		wantValid bool
	}{
		{name: "komentar 10 karakter lolos", comment: "sepuluh ka", wantValid: true},
		{name: "komentar 9 karakter ditolak", comment: "sembilan!", wantValid: false},
		{name: "komentar kosong ditolak", comment: "", wantValid: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fields := helper.ValidateStruct(model.CreateReviewRequest{
				Rating:  4,
				Comment: tc.comment,
			})
			got := len(fields) == 0
			if got != tc.wantValid {
				t.Fatalf("komentar %q valid = %v, mau %v (fields: %v)", tc.comment, got, tc.wantValid, fields)
			}
		})
	}
}

func TestReviewTagRules(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		tags      []string
		wantValid bool
	}{
		{name: "tanpa tag lolos", tags: nil, wantValid: true},
		{name: "satu tag dua karakter lolos", tags: []string{"go"}, wantValid: true},
		{name: "tag satu karakter ditolak", tags: []string{"g"}, wantValid: false},
		{name: "tag 51 karakter ditolak", tags: []string{strings.Repeat("x", 51)}, wantValid: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fields := helper.ValidateStruct(model.CreateReviewRequest{
				Rating:  4,
				Comment: "Komentar yang cukup panjang untuk lolos.",
				Tags:    tc.tags,
			})
			got := len(fields) == 0
			if got != tc.wantValid {
				t.Fatalf("tags %v valid = %v, mau %v (fields: %v)", tc.tags, got, tc.wantValid, fields)
			}
		})
	}
}

// TestReviewTagsAreNormalized memastikan tag ulasan disimpan dalam bentuk
// ternormalisasi: huruf kecil, spasi dipangkas, duplikat dibuang.
func TestReviewTagsAreNormalized(t *testing.T) {
	t.Parallel()

	got := NormalizeTags([]string{" Seru ", "seru", "Go", "  ", "SQL"})
	want := []string{"seru", "go", "sql"}

	assertTags(t, got, want)
}
