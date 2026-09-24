package service

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// translateError menerjemahkan error repository menjadi *helper.AppError
// agar ErrorHandler terpusat yang membentuk JSON.
//
// Pola pakai:
//
//	_, err := s.Repo.GetByID(...)
//	if err != nil {
//	    return translateError(err, "mahasiswa")
//	}
//
// Aturan:
//
//	repository.ErrNotFound / ErrStudentNotFound -> helper.NotFound(resource)
//	repository.ErrDuplicate                    -> helper.Conflict("data sudah ada")
//	pg 23505 (unique)                          -> helper.Conflict("data sudah ada")
//	pg 23503 (foreign key)                     -> helper.BadRequest("referensi tidak valid")
//	pg 23502 (not null)                        -> helper.BadRequest("field wajib tidak diisi")
//	pg 23514 (check)                           -> helper.BadRequest("nilai tidak valid")
//	default                                    -> helper.Internal(err) (fail closed)
//
// Default tidak pernah mengembalikan nil (Modul 7 bug #6).
func translateError(err error, resource string) *helper.AppError {
	if errors.Is(err, repository.ErrNotFound) ||
		errors.Is(err, repository.ErrStudentNotFound) {
		return helper.NotFound(resource)
	}
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Conflict("data sudah ada")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return helper.Conflict("data sudah ada")
		case "23503":
			return helper.BadRequest("referensi tidak valid")
		case "23502":
			return helper.BadRequest("field wajib tidak diisi")
		case "23514":
			return helper.BadRequest("nilai tidak valid")
		}
	}
	return helper.Internal(err)
}
