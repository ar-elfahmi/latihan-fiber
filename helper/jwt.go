package helper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DefaultTokenTTL adalah masa berlaku access token (12 jam), sama seperti
// yang dipakai sebelum penerbitan token dipindah ke package ini.
const DefaultTokenTTL = 12 * time.Hour

// JWTManager menerbitkan dan memverifikasi access token JWT.
//
// Klaim token SENGAJA tidak membawa daftar permission. Sumber kebenaran
// tunggal permission adalah database (middleware.Protect query live di
// setiap request). Alasan:
//   - perubahan role harus segera berlaku walau token masih hidup;
//   - token tidak membesar & tidak ada risiko perms stale;
//   - middleware tetap fail closed: bilamana DB tidak bisa dibaca -> 403.
//
// File ini tidak mengimpor fiber maupun app/service supaya bisa dipakai
// di lapisan mana pun tanpa siklus import.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTManager membuat penerbit token dengan kunci tanda tangan dan masa
// berlaku bawaan (DefaultTokenTTL).
func NewJWTManager(secret []byte) *JWTManager {
	return &JWTManager{secret: secret, ttl: DefaultTokenTTL}
}

// GenerateAccessToken membuat token JWT berisi id + username, berlaku
// selama ttl. Kunci kosong ditolak (fail closed).
func (m *JWTManager) GenerateAccessToken(userID int, username string) (string, error) {
	if m == nil || len(m.secret) == 0 {
		return "", errors.New("JWT secret belum dikonfigurasi")
	}
	claims := jwt.MapClaims{
		"sub":      userID,                       // siapa pemilik token (id user)
		"username": username,                     // nama user, biar gampang dibaca
		"exp":      time.Now().Add(m.ttl).Unix(), // waktu kedaluwarsa
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// ParseToken memeriksa tanda tangan + kedaluwarsa token, lalu mengembalikan
// isinya. Error berarti token palsu, rusak, atau sudah kedaluwarsa.
func (m *JWTManager) ParseToken(tokenString string) (jwt.MapClaims, error) {
	if m == nil || len(m.secret) == 0 {
		return nil, errors.New("JWT secret belum dikonfigurasi")
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		// Tolak algoritma lain (misal "none") yang bisa dipakai menipu.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("algoritma token tidak dikenal")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("token tidak valid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("isi token tidak dikenal")
	}
	return claims, nil
}
