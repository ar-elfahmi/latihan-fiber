package helper

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestRefreshTokenRotation memastikan refresh token hanya bisa dipakai sekali:
// consume pertama sukses, consume kedua (replay) ditolak.
func TestRefreshTokenRotation(t *testing.T) {
	m := NewJWTManager([]byte("secret-uji-refresh"))

	token, err := m.GenerateRefreshToken(7, "budi")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	claims, err := m.ConsumeRefreshToken(token)
	if err != nil {
		t.Fatalf("consume pertama: %v", err)
	}
	if claims["username"] != "budi" {
		t.Errorf("username = %v, mau budi", claims["username"])
	}
	if _, err := m.ConsumeRefreshToken(token); err == nil {
		t.Error("consume kedua (replay) harus ditolak")
	}
}

// TestRefreshTokenRejectsAccessToken memastikan access token tidak bisa
// dipakai sebagai refresh token.
func TestRefreshTokenRejectsAccessToken(t *testing.T) {
	m := NewJWTManager([]byte("secret-uji-refresh"))

	access, err := m.GenerateAccessToken(7, "budi")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	if _, err := m.ConsumeRefreshToken(access); err == nil {
		t.Error("access token harus ditolak saat dipakai sebagai refresh token")
	}
	if err := m.RevokeRefreshToken(access); err == nil {
		t.Error("revoke access token harus ditolak")
	}
}

// TestRevokeRefreshToken memastikan logout mencabut token: revoke sukses,
// consume sesudah revoke ditolak, revoke token rusak ditolak.
func TestRevokeRefreshToken(t *testing.T) {
	m := NewJWTManager([]byte("secret-uji-refresh"))

	token, err := m.GenerateRefreshToken(7, "budi")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if err := m.RevokeRefreshToken(token); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := m.ConsumeRefreshToken(token); err == nil {
		t.Error("consume setelah revoke harus ditolak")
	}
	if err := m.RevokeRefreshToken("bukan-token"); err == nil {
		t.Error("revoke token rusak harus ditolak")
	}
}

// TestGenerateRefreshTokenRejectsEmptySecret memastikan manager tanpa
// secret fail closed untuk refresh token.
func TestGenerateRefreshTokenRejectsEmptySecret(t *testing.T) {
	m := NewJWTManager(nil)
	if _, err := m.GenerateRefreshToken(1, "budi"); err == nil {
		t.Error("tanpa secret, penerbitan refresh token harus gagal")
	}
}

// TestRefreshTokenExpiry memastikan refresh token kedaluwarsa ditolak
// (token dibuat langsung dengan exp lampau; tanpa tidur di test).
func TestRefreshTokenExpiry(t *testing.T) {
	secret := []byte("secret-uji-refresh")
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      7,
		"username": "budi",
		"typ":      "refresh",
		"exp":      time.Now().Add(-time.Minute).Unix(),
	})
	token, err := expired.SignedString(secret)
	if err != nil {
		t.Fatalf("penandatanganan token uji gagal: %v", err)
	}
	if _, err := NewJWTManager(secret).ConsumeRefreshToken(token); err == nil {
		t.Error("refresh token kedaluwarsa harus ditolak")
	}
}
