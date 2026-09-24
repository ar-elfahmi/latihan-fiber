package service

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// AuthService mengurus login: cek akun di tabel users, lalu terbitkan token JWT.
// Modul 6: permission tidak disisipkan ke token; middleware.Protect akan
// mengambilnya langsung dari tabel RBAC saat request berikutnya.
type AuthService struct {
	Repo   repository.UserRepository
	RBAC   repository.RBACRepository // saat ini unused; disimpan untuk kompatibilitas wiring
	Secret []byte // kunci tanda tangan token, dibaca dari JWT_SECRET di .env
}

// Login memeriksa username + password, lalu memberi token JWT.
// Alur: baca body -> validasi -> cari user -> cocokkan password -> buat token.
func (s AuthService) Login(c *fiber.Ctx) error {
	// 1. Baca body JSON.
	var req model.LoginRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.Error(c, 400, "body harus berupa JSON yang valid")
	}

	// 2. Validasi bentuk input (kosong/tidak).
	if errs := ValidateLogin(req); len(errs) > 0 {
		return c.Status(422).JSON(fiber.Map{"error": "validasi gagal", "fields": errs})
	}

	// 3. Cari akun. Salah username ATAU salah password: pesannya sama,
	//    supaya orang lain tidak tahu field mana yang keliru.
	u, err := s.Repo.GetByUsername(c.Context(), req.Username)
	if err != nil || !CheckPassword(req.Password, u.PasswordHash) {
		return helper.Error(c, 401, "username atau password salah")
	}

	// 4. Buat token dan kirimkan ke client.
	token, err := MakeToken(u.ID, u.Username, s.Secret)
	if err != nil {
		return helper.Error(c, 500, "gagal membuat token")
	}
	return c.JSON(fiber.Map{
		"token": token,
		"user":  fiber.Map{"id": u.ID, "username": u.Username},
	})
}

// CheckPassword membandingkan password asli dengan hash bcrypt di database.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
