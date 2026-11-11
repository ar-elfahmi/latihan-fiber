package helper

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DefaultTokenTTL adalah masa berlaku access token (12 jam), sama seperti
// yang dipakai sebelum penerbitan token dipindah ke package ini.
const DefaultTokenTTL = 12 * time.Hour

// DefaultRefreshTTL adalah masa berlaku refresh token (7 hari).
// Disepakati dengan solo dev (S-004 Q2); nilai final bisa dioverride lewat
// konfigurasi bila dibutuhkan tanpa mengubah kode.
const DefaultRefreshTTL = 7 * 24 * time.Hour

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

	refreshTTL time.Duration

	// Denylist refresh token yang sudah dipakai/dicabut (rotation).
	// In-memory: hilang saat restart — diterima dan didokumentasikan
	// (S-004 Q2). Access token tetap stateless: logout hanya mencabut
	// refresh token, access token valid sampai kedaluwarsa.
	mu      sync.Mutex
	revoked map[string]struct{}
}

// NewJWTManager membuat penerbit token dengan kunci tanda tangan dan masa
// berlaku bawaan (DefaultTokenTTL).
func NewJWTManager(secret []byte) *JWTManager {
	return &JWTManager{secret: secret, ttl: DefaultTokenTTL, refreshTTL: DefaultRefreshTTL, revoked: map[string]struct{}{}}
}

// SetRefreshTTL mengubah masa berlaku refresh token (untuk test/konfigurasi).
func (m *JWTManager) SetRefreshTTL(d time.Duration) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshTTL = d
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

// GenerateRefreshToken membuat refresh token JWT (klaim typ=refresh + jti acak
// supaya tiap token unik dan bisa dicabut satu per satu).
func (m *JWTManager) GenerateRefreshToken(userID int, username string) (string, error) {
	if m == nil || len(m.secret) == 0 {
		return "", errors.New("JWT secret belum dikonfigurasi")
	}
	m.mu.Lock()
	ttl := m.refreshTTL
	if ttl <= 0 {
		ttl = DefaultRefreshTTL
	}
	m.mu.Unlock()
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"typ":      "refresh",
		"jti":      time.Now().UTC().UnixNano(),
		"exp":      time.Now().Add(ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// ConsumeRefreshToken memvalidasi refresh token lalu mencabutnya (rotation):
// token lama tidak bisa dipakai dua kali, pemakaian ulang → 401.
func (m *JWTManager) ConsumeRefreshToken(tokenString string) (jwt.MapClaims, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims["typ"] != "refresh" {
		return nil, errors.New("bukan refresh token")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.revoked == nil {
		m.revoked = map[string]struct{}{}
	}
	if _, bad := m.revoked[tokenString]; bad {
		return nil, errors.New("refresh token sudah dicabut")
	}
	m.revoked[tokenString] = struct{}{}
	return claims, nil
}

// RevokeRefreshToken mencabut refresh token tanpa menerbitkan pengganti
// (dipakai POST /auth/logout). Token yang tidak valid tetap ditolak.
func (m *JWTManager) RevokeRefreshToken(tokenString string) error {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return err
	}
	if claims["typ"] != "refresh" {
		return errors.New("bukan refresh token")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.revoked == nil {
		m.revoked = map[string]struct{}{}
	}
	m.revoked[tokenString] = struct{}{}
	return nil
}
