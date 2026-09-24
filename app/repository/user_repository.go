package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		`SELECT id, username, password_hash, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("cari user: %w", err)
	}
	return u, nil
}

// List mengambil halaman pertama user dengan urutan created_at DESC,
// id DESC. Limit +1 dipakai untuk deteksi has_more di service.
func (r UserRepository) List(ctx context.Context, limit int) ([]model.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, username, password_hash, created_at
		FROM users
		ORDER BY created_at DESC, id DESC
		LIMIT $1
	`, limit+1)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

// FindAfterCursor mengambil daftar user setelah (cursor, cursorID) dengan
// urutan yang sama. Digunakan service untuk halaman cursor.
//
// Arah ORDER BY harus cocok dengan index users_created_at_id_desc_idx
// (DESC). Kalau terbalik, planner memilih Seq Scan (Modul 7 bug #8).
func (r UserRepository) FindAfterCursor(ctx context.Context, cursor time.Time, cursorID, limit int) ([]model.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE (created_at, id) < ($1, $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`, cursor, cursorID, limit+1)
	if err != nil {
		return nil, fmt.Errorf("cursor users: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

// scanUsers membaca semua baris menjadi slice model.User. Dipakai
// oleh List dan FindAfterCursor agar tidak duplikasi.
func scanUsers(rows pgx.Rows) ([]model.User, error) {
	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi user: %w", err)
	}
	return users, nil
}
