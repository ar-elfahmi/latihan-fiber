package helper

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAccessTokenIsValidForSameManager(t *testing.T) {
	manager := NewJWTManager([]byte("secret-uji"))

	token, err := manager.GenerateAccessToken(7, "budi")
	if err != nil {
		t.Fatalf("GenerateAccessToken gagal: %v", err)
	}

	claims, err := manager.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken gagal untuk token sendiri: %v", err)
	}
	if claims["username"] != "budi" {
		t.Errorf("username harus budi, dapat %v", claims["username"])
	}
	if sub, ok := claims["sub"].(float64); !ok || int(sub) != 7 {
		t.Errorf("sub harus 7, dapat %v", claims["sub"])
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("exp harus ada, dapat %v", claims["exp"])
	}
	if exp <= float64(time.Now().Unix()) {
		t.Error("token yang baru terbit harus belum kedaluwarsa")
	}
}

func TestParseTokenRejectsForeignSignature(t *testing.T) {
	issuer := NewJWTManager([]byte("secret-uji"))
	verifier := NewJWTManager([]byte("secret-lain"))

	token, err := issuer.GenerateAccessToken(1, "siti")
	if err != nil {
		t.Fatalf("GenerateAccessToken gagal: %v", err)
	}
	if _, err := verifier.ParseToken(token); err == nil {
		t.Error("token dengan tanda tangan lain harus ditolak")
	}
}

func TestParseTokenRejectsGarbage(t *testing.T) {
	manager := NewJWTManager([]byte("secret-uji"))
	if _, err := manager.ParseToken("bukan-token"); err == nil {
		t.Error("token rusak harus ditolak")
	}
}

func TestParseTokenRejectsExpired(t *testing.T) {
	manager := NewJWTManager([]byte("secret-uji"))
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      1,
		"username": "lama",
		"exp":      time.Now().Add(-time.Minute).Unix(),
	})
	token, err := expired.SignedString([]byte("secret-uji"))
	if err != nil {
		t.Fatalf("penandatanganan token uji gagal: %v", err)
	}
	if _, err := manager.ParseToken(token); err == nil {
		t.Error("token kedaluwarsa harus ditolak")
	}
}

func TestJWTManagerWithoutSecretFailsClosed(t *testing.T) {
	manager := NewJWTManager(nil)
	if _, err := manager.GenerateAccessToken(1, "budi"); err == nil {
		t.Error("tanpa secret, penerbitan token harus gagal")
	}
	if _, err := manager.ParseToken("apa saja"); err == nil {
		t.Error("tanpa secret, parsing token harus gagal")
	}
}
