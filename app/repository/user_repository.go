package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
)

// UserRepository membaca akun login dari tabel users.
type UserRepository struct {
	// DB adalah connection pool yang sama dengan repository lain.
	DB *pgxpool.Pool
}

// GetByUsername mengambil satu akun berdasarkan username.
// Mengembalikan ErrNotFound bila username tidak terdaftar.
func (r UserRepository) GetByUsername(ctx context.Context, username string) (model.User, error) {
	var u model.User
	err := r.DB.QueryRow(ctx,
		`SELECT id, username, password_hash FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("cari user: %w", err)
	}
	return u, nil
}
