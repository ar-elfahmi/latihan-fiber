package model

import "time"

// User adalah baris pada tabel users (akun login).
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // hash tidak pernah dikirim ke client
	Email        string    `json:"email,omitempty"`
	IsActive     bool      `json:"is_active,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// CursorMeta adalah metadata pagination cursor untuk ListUserResponse.
// Tanpa Total/TotalPages: cursor pagination tidak menjamin jumlah
// total baris karena OFFSET tidak dipakai. Client cukup mengikuti
// next_cursor hingga has_more=false.
type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// AuthUser adalah identitas pemanggil yang sudah terverifikasi.
// Diisi middleware.Protect dari klaim token, lalu dibaca service
// untuk pemeriksaan kepemilikan (dipakai CanAccessStudent).
type AuthUser struct {
	ID       int
	Username string
}

// LoginRequest adalah body POST /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateUserRequest adalah body POST /users (admin membuat akun baru).
// Validasi sepenuhnya lewat tag validate; tidak ada pemeriksaan manual
// tambahan di service.
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace"`
	IsActive bool   `json:"is_active"`
}

// ReplaceUserRequest adalah body PUT /users/:id (ganti seluruh field).
type ReplaceUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace"`
	IsActive bool   `json:"is_active"`
}

// PatchUserRequest adalah body PATCH /users/:id (ubah sebagian).
// Field bertipe pointer dan tag omitnil: field yang tidak dikirim
// (nil) dilewati, field yang dikirim divalidasi sesuai tag.
type PatchUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3,max=30,alphanum"`
	Email    *string `json:"email,omitempty"    validate:"omitnil,email,max=120"`
	Password *string `json:"password,omitempty" validate:"omitnil,min=8,max=72,nospace"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// RegisterRequest adalah body POST /auth/register (publik, mendaftarkan
// diri sendiri). Password di sini menggunakan tag strongpassword:
// minimal 8 karakter, harus ada huruf besar, huruf kecil, dan angka.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,max=72,strongpassword"`
}

// ListUserQuery membawa parameter query string untuk GET /users.
type ListUserQuery struct {
	Page  int    // ?page=
	Limit int    // ?limit=
	Sort  string // ?sort=
	Order string // ?order=
}

// Offset dipakai di query SQL LIMIT/OFFSET.
func (q ListUserQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
