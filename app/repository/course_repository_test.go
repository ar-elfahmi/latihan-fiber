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
