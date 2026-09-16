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
