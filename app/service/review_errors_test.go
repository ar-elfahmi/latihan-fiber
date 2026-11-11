package service

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// translateReviewError memetakan error repository ulasan menjadi
// *helper.AppError. File ini WAJIB package service (bukan service_test)
// karena translateReviewError tidak diekspor (temuan B1 #2).

func TestTranslateReviewError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "id tidak valid -> 400 BAD_REQUEST",
			err:        repository.ErrInvalidID,
			wantStatus: http.StatusBadRequest,
			wantCode:   helper.CodeBadRequest,
		},
		{
			name:       "kelas tidak ada -> 404 NOT_FOUND",
			err:        repository.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   helper.CodeNotFound,
		},
		{
			name:       "baris tidak ada -> 404 NOT_FOUND",
			err:        pgx.ErrNoRows,
			wantStatus: http.StatusNotFound,
			wantCode:   helper.CodeNotFound,
		},
		{
			name:       "ulasan duplikat -> 409 ALREADY_REVIEWED",
			err:        repository.ErrDuplicate,
			wantStatus: http.StatusConflict,
			wantCode:   "ALREADY_REVIEWED",
		},
		{
			name:       "bukan pemilik -> 403 FORBIDDEN",
			err:        repository.ErrNotOwner,
			wantStatus: http.StatusForbidden,
			wantCode:   helper.CodeForbidden,
		},
		{
			name:       "menilai ulasan sendiri -> 409 OWN_REVIEW",
			err:        repository.ErrOwnReview,
			wantStatus: http.StatusConflict,
			wantCode:   "OWN_REVIEW",
		},
		{
			name:       "sudah menilai -> 409 ALREADY_VOTED",
			err:        repository.ErrAlreadyVoted,
			wantStatus: http.StatusConflict,
			wantCode:   "ALREADY_VOTED",
		},
		{
			name:       "batas balasan -> 409 REPLY_LIMIT",
			err:        repository.ErrReplyLimit,
			wantStatus: http.StatusConflict,
			wantCode:   "REPLY_LIMIT",
		},
		{
			name:       "error tak dikenal -> 500 INTERNAL",
			err:        errors.New("kegagalan tak terduga"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   helper.CodeInternal,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := translateReviewError(tc.err)
			appErr := helper.AsAppError(got)
			if appErr == nil {
				t.Fatalf("translateReviewError(%v) bukan *helper.AppError: %T", tc.err, got)
			}
			if appErr.Status != tc.wantStatus || appErr.Code != tc.wantCode {
				t.Fatalf("translateReviewError(%v) = %d %s, mau %d %s",
					tc.err, appErr.Status, appErr.Code, tc.wantStatus, tc.wantCode)
			}
		})
	}
}

// TestTranslateReviewErrorWrapped memastikan error yang dibungkus (wrapped)
// tetap dipetakan dengan benar: errors.Is, bukan perbandingan nil.
func TestTranslateReviewErrorWrapped(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("menulis ulasan: %w", repository.ErrDuplicate)
	appErr := helper.AsAppError(translateReviewError(wrapped))
	if appErr == nil {
		t.Fatal("translateReviewError(wrapped) bukan *helper.AppError")
	}
	if appErr.Status != http.StatusConflict || appErr.Code != "ALREADY_REVIEWED" {
		t.Fatalf("hasil = %d %s, mau 409 ALREADY_REVIEWED", appErr.Status, appErr.Code)
	}
}

// TestReviewValidationCodeIsStable mengunci kode error validasi.
// Kode stabil adalah VALIDATION_FAILED (helper.CodeValidationFailed, D-002),
// BUKAN "VALIDATION_ERROR" seperti tertulis di modul (temuan B1 #9).
func TestReviewValidationCodeIsStable(t *testing.T) {
	t.Parallel()

	if helper.CodeValidationFailed != "VALIDATION_FAILED" {
		t.Fatalf("helper.CodeValidationFailed = %q, mau %q", helper.CodeValidationFailed, "VALIDATION_FAILED")
	}

	appErr := helper.Validation(map[string]string{"rating": "nilai minimal 1"})
	if appErr.Code != helper.CodeValidationFailed {
		t.Fatalf("kode validasi = %q, mau %q", appErr.Code, helper.CodeValidationFailed)
	}
	if appErr.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status validasi = %d, mau %d", appErr.Status, http.StatusUnprocessableEntity)
	}
}
