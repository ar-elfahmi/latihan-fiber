package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Error domain yang dipakai seluruh repository.
// Service menerjemahkan error ini menjadi status HTTP.
var (
	ErrNotFound        = errors.New("data tidak ditemukan")
	ErrDuplicate       = errors.New("data sudah ada")
	ErrStudentNotFound = errors.New("mahasiswa tidak ditemukan")
	// Sentinel ulasan kelas (MongoDB). Dipakai ReviewService lewat
	// translateReviewError (S-004); bentuk tabel modul, S-005 white-box test.
	ErrInvalidID    = errors.New("id tidak valid")
	ErrNotOwner     = errors.New("bukan pemilik data")
	ErrOwnReview    = errors.New("tidak boleh menilai ulasan sendiri")
	ErrAlreadyVoted = errors.New("sudah menilai ulasan ini")
	ErrReplyLimit   = errors.New("jumlah balasan sudah mencapai batas")
)

// isDuplicate mendeteksi pelanggaran unique (kode PG 23505).
func isDuplicate(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// isForeignKey mendeteksi pelanggaran foreign key (kode PG 23503).
func isForeignKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
