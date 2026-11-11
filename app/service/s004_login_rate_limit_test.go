package service

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/config"
	"latihan-fiber/helper"
)

// s004FakeUsers adalah fake repository.UserRepository khusus slice S-004.
// Fake lokal per test (tanpa state global) — docs/TESTING.md §4.
type s004FakeUsers struct {
	repository.UserRepository // method lain: panic bila terpanggil

	user model.User
	err  error
}

func (f *s004FakeUsers) GetByUsername(_ context.Context, _ string) (model.User, error) {
	if f.err != nil {
		return model.User{}, f.err
	}
	return f.user, nil
}

var _ repository.UserRepository = (*s004FakeUsers)(nil)

// s004ServeLogin memanggil AuthService.Login lewat route Fiber sungguhan
// supaya IP (untuk rate limit) terbaca seperti di production.
func s004ServeLogin(t *testing.T, svc AuthService, body string) (int, map[string]any) {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
	app.Post("/auth/login", svc.Login)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	var parsed map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("POST /auth/login: response bukan JSON: %v", err)
	}
	return resp.StatusCode, parsed
}

// s004AuthService menyusun AuthService dengan fake user + limiter sendiri.
func s004AuthService(t *testing.T, max int) AuthService {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password-benar"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	users := &s004FakeUsers{user: model.User{ID: 3, Username: "budi", PasswordHash: string(hash)}}
	return AuthService{
		Repo:    users,
		Tokens:  helper.NewJWTManager([]byte("secret-uji-s004")),
		Limiter: helper.NewLoginLimiter(max, time.Minute),
	}
}

// TestS004LoginRateLimit memastikan login dibatasi 5 kali gagal per menit
// per IP: percobaan keenam dan seterusnya dijawab 429.
func TestS004LoginRateLimit(t *testing.T) {
	t.Parallel()
	svc := s004AuthService(t, helper.DefaultLoginLimit)
	body := `{"username":"budi","password":"password-salah"}`

	for i := 0; i < helper.DefaultLoginLimit; i++ {
		status, parsed := s004ServeLogin(t, svc, body)
		if status != 401 {
			t.Fatalf("percobaan ke-%d: status = %d, mau 401 (body: %v)", i+1, status, parsed)
		}
	}
	status, parsed := s004ServeLogin(t, svc, body)
	if status != 429 || parsed["code"] != helper.CodeTooManyRequests {
		t.Errorf("percobaan ke-%d: dapat %d %v, mau 429 %s",
			helper.DefaultLoginLimit+1, status, parsed["code"], helper.CodeTooManyRequests)
	}
}

// TestS004LoginSuccessResetsRateLimit memastikan login berhasil me-reset
// penghitung kegagalan sehingga IP tidak terkunci.
func TestS004LoginSuccessResetsRateLimit(t *testing.T) {
	t.Parallel()
	svc := s004AuthService(t, 2)

	if status, parsed := s004ServeLogin(t, svc, `{"username":"budi","password":"salah"}`); status != 401 {
		t.Fatalf("kegagalan login: status = %d, mau 401 (body: %v)", status, parsed)
	}
	if status, parsed := s004ServeLogin(t, svc, `{"username":"budi","password":"password-benar"}`); status != 200 {
		t.Fatalf("login benar: status = %d, mau 200 (body: %v)", status, parsed)
	}
	// Setelah login berhasil, IP harus bebas mencoba lagi (401, bukan 429).
	if status, parsed := s004ServeLogin(t, svc, `{"username":"budi","password":"salah"}`); status != 401 {
		t.Errorf("percobaan setelah reset: status = %d, mau 401 (bukan 429) (body: %v)", status, parsed)
	}
}
