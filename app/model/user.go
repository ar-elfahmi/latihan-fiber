package model

// User adalah baris pada tabel users (akun login).
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // hash tidak pernah dikirim ke client
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
