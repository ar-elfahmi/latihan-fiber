package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Bagian ini adalah "kamus sandi" JWT:
// - MakeToken  : membungkus data user menjadi token.
// - ParseToken : membaca token dan memastikan gelang itu benar-benar buatan
//                kita (tanda tangan cocok) dan belum kedaluwarsa.
//
// Modul 6: klaim TIDAK membawa daftar permission. Sumber kebenaran tunggal
// adalah database (Protect query live di setiap request). Alasan:
//   - perubahan role harus segera berlaku walau token masih hidup;
//   - token tidak membesar & tidak ada risiko perms stale;
//   - middleware tetap fail closed: bilamana DB tidak bisa dibaca -> 403.

// MakeToken membuat token JWT berisi id + username, berlaku 12 jam.
func MakeToken(userID int, username string, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID,                                 // siapa pemilik token (id user)
		"username": username,                               // nama user, biar gampang dibaca
		"exp":      time.Now().Add(12 * time.Hour).Unix(), // waktu kedaluwarsa
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// ParseToken memeriksa tanda tangan + kedaluwarsa token, lalu mengembalikan isinya.
// Error berarti token palsu, rusak, atau sudah kedaluwarsa.
func ParseToken(tokenString string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		// Tolak algoritma lain (misal "none") yang bisa dipakai menipu.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("algoritma token tidak dikenal")
		}
		return secret, nil
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
