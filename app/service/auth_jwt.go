package service

import (
	"github.com/golang-jwt/jwt/v5"

	"latihan-fiber/helper"
)

// "Kamus sandi" JWT:
// - helper.JWTManager : menerbitkan token (dipakai service & main.go).
// - ParseToken        : membaca token dan memastikan gelang itu benar-benar
//                       buatan kita (tanda tangan cocok) dan belum kedaluwarsa.
//
// Modul 6: klaim TIDAK membawa daftar permission. Sumber kebenaran tunggal
// adalah database (Protect query live di setiap request). Alasan:
//   - perubahan role harus segera berlaku walau token masih hidup;
//   - token tidak membesar & tidak ada risiko perms stale;
//   - middleware tetap fail closed: bilamana DB tidak bisa dibaca -> 403.
//
// Isi token (sub, username, exp) sama seperti sebelumnya pemindahan
// penerbitan token ke helper, jadi token lama tetap sah untuk Protect.

// ParseToken memeriksa tanda tangan + kedaluwarsa token, lalu mengembalikan isinya.
// Error berarti token palsu, rusak, atau sudah kedaluwarsa.
func ParseToken(tokenString string, secret []byte) (jwt.MapClaims, error) {
	return helper.NewJWTManager(secret).ParseToken(tokenString)
}
// claimIdentity membaca sub (id user) dan username dari klaim refresh token.
func claimIdentity(claims map[string]any) (int, string) {
	id := 0
	switch v := claims["sub"].(type) {
	case float64:
		id = int(v)
	case int:
		id = v
	case int64:
		id = int(v)
	}
	username, _ := claims["username"].(string)
	return id, username
}
