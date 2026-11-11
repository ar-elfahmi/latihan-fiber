package repository

import (
	"errors"
	"testing"

	"latihan-fiber/app/model"
)

// stubRules mengembalikan aturan domain yang selalu terisi, untuk dipakai
// pada kasus yang tidak menguji aturan kosong.
func stubRules() CreditRules {
	return CreditRules{
		WithinCreditLimit: func(current, delta int) bool {
			return current+delta <= model.MaxCreditsPerSemester
		},
		CreditsAfterSwap: func(currentCredits int, oldCourse, newCourse model.Course) int {
			return currentCredits - oldCourse.Credits + newCourse.Credits
		},
	}
}

// TestNewPostgresCourseRepository memastikan repository KRS hanya bisa dibuat
// dengan aturan batas SKS yang lengkap: repository tidak boleh dipakai tanpa
// aturan (kalau tidak, pemeriksaannya bisa dilewati diam-diam).
func TestNewPostgresCourseRepository(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		rules   CreditRules
		wantErr bool
	}{
		{
			name:    "aturan lengkap -> repository dibuat",
			rules:   stubRules(),
			wantErr: false,
		},
		{
			name:    "tanpa WithinCreditLimit -> ditolak",
			rules:   CreditRules{CreditsAfterSwap: stubRules().CreditsAfterSwap},
			wantErr: true,
		},
		{
			name:    "tanpa CreditsAfterSwap -> ditolak",
			rules:   CreditRules{WithinCreditLimit: stubRules().WithinCreditLimit},
			wantErr: true,
		},
		{
			name:    "tanpa aturan sama sekali -> ditolak",
			rules:   CreditRules{},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, err := NewPostgresCourseRepository(nil, tc.rules)
			if tc.wantErr {
				if !errors.Is(err, ErrMissingCreditLimitRules) {
					t.Fatalf("err = %v, mau %v", err, ErrMissingCreditLimitRules)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPostgresCourseRepository: %v", err)
			}
			if repo.rules.WithinCreditLimit == nil || repo.rules.CreditsAfterSwap == nil {
				t.Fatal("aturan domain KRS belum tersimpan di repository")
			}
		})
	}
}

// TestAfterSwapWithinLimitMemakaiAturanTerinjeksi membuktikan jalur tukar
// kelas memakai aturan yang diinjeksi, bukan membandingkan langsung dengan
// angka batas di kode. Aturan di sini memakai batas 12 SKS: kalau jalur Swap
// masih memakai model.MaxCreditsPerSemester (24), kasus terakhir akan lolos
// dan test ini gagal.
func TestAfterSwapWithinLimitMemakaiAturanTerinjeksi(t *testing.T) {
	t.Parallel()

	rules := CreditRules{
		WithinCreditLimit: func(current, delta int) bool { return current+delta <= 12 },
		CreditsAfterSwap: func(currentCredits int, oldCourse, newCourse model.Course) int {
			return currentCredits - oldCourse.Credits + newCourse.Credits
		},
	}
	repo, err := NewPostgresCourseRepository(nil, rules)
	if err != nil {
		t.Fatalf("NewPostgresCourseRepository: %v", err)
	}

	from := model.Course{Credits: 3}
	to := model.Course{Credits: 4}

	cases := []struct {
		name        string
		totalBefore int
		want        bool
	}{
		{name: "SKS lama 9 -> setelah tukar 10, lolos untuk batas 12", totalBefore: 9, want: true},
		{name: "SKS lama 11 -> setelah tukar 12, lolos tepat di batas 12", totalBefore: 11, want: true},
		{name: "SKS lama 12 -> setelah tukar 13, ditolak untuk batas 12", totalBefore: 12, want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := repo.afterSwapWithinLimit(tc.totalBefore, from, to)
			if got != tc.want {
				t.Fatalf("afterSwapWithinLimit(%d, 3 SKS, 4 SKS) = %v, mau %v",
					tc.totalBefore, got, tc.want)
			}
		})
	}
}
