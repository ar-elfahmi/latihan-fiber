package service

import (
	"errors"

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
	Tokens *helper.JWTManager        // penerbit token JWT (helper.NewJWTManager)
}

func (s AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	if len(req.Password) > 72 {
		return helper.Validation(map[string]string{"password": "maksimal 72 byte"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return helper.Internal(err)
	}
	id, err := s.Repo.Register(c.UserContext(), req, string(hash))
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Conflict("username sudah digunakan")
	}
	if err != nil {
		return helper.Internal(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "username": req.Username, "email": req.Email})
}

// Login memeriksa username + password, lalu memberi token JWT.
// Alur: baca body -> validasi -> cari user -> cocokkan password -> buat token.
func (s AuthService) Login(c *fiber.Ctx) error {
	// 1. Baca body JSON.
	var req model.LoginRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	// 2. Validasi bentuk input (kosong/tidak).
	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	// 3. Cari akun. Salah username ATAU salah password: pesannya sama,
	//    supaya orang lain tidak tahu field mana yang keliru.
	u, err := s.Repo.GetByUsername(c.Context(), req.Username)
	if err != nil || !CheckPassword(req.Password, u.PasswordHash) {
		return helper.Unauthorized("username atau password salah")
	}

	// 4. Buat token dan kirimkan ke client.
	if s.Tokens == nil {
		return helper.Internal(errors.New("penerbit token JWT belum dikonfigurasi"))
	}
	token, err := s.Tokens.GenerateAccessToken(u.ID, u.Username)
	if err != nil {
		return helper.Internal(err)
	}
	// Non-breaking (S-004 Q1): respons lama {token, user} dipertahankan,
	// refresh_token ditambahkan sebagai field baru.
	refresh, err := s.Tokens.GenerateRefreshToken(u.ID, u.Username)
	if err != nil {
		return helper.Internal(err)
	}
	return c.JSON(fiber.Map{
		"token":         token,
		"refresh_token": refresh,
		"user":          fiber.Map{"id": u.ID, "username": u.Username},
	})
}

// Refresh menukar refresh token dengan pasangan token baru (rotation):
// token lama langsung dicabut, pemakaian ulang → 401.
func (s AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	if s.Tokens == nil {
		return helper.Internal(errors.New("penerbit token JWT belum dikonfigurasi"))
	}
	claims, err := s.Tokens.ConsumeRefreshToken(req.RefreshToken)
	if err != nil {
		return helper.Unauthorized("refresh token tidak valid, kedaluwarsa, atau sudah dicabut")
	}
	userID, username := claimIdentity(claims)
	if userID <= 0 {
		return helper.Unauthorized("refresh token tidak berisi identitas yang sah")
	}
	access, err := s.Tokens.GenerateAccessToken(userID, username)
	if err != nil {
		return helper.Internal(err)
	}
	refresh, err := s.Tokens.GenerateRefreshToken(userID, username)
	if err != nil {
		return helper.Internal(err)
	}
	return c.JSON(model.WebResponse{
		Success: true,
		Message: "token berhasil diperbarui",
		Data:    model.TokenPair{AccessToken: access, RefreshToken: refresh},
	})
}

// Logout mencabut refresh token. Access token tetap stateless: tetap valid
// sampai kedaluwarsa (didokumentasikan, S-004 Q2).
func (s AuthService) Logout(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	if s.Tokens == nil {
		return helper.Internal(errors.New("penerbit token JWT belum dikonfigurasi"))
	}
	if err := s.Tokens.RevokeRefreshToken(req.RefreshToken); err != nil {
		return helper.Unauthorized("refresh token tidak valid, kedaluwarsa, atau sudah dicabut")
	}
	return c.JSON(model.WebResponse{Success: true, Message: "berhasil keluar"})
}

// Me mengembalikan profil user yang sedang login (tanpa password_hash —
// model.User menandai field itu json:"-").
func (s AuthService) Me(c *fiber.Ctx) error {
	u, err := s.Repo.FindByID(c.UserContext(), CurrentUserID(c))
	if err != nil {
		return translateError(err, "user")
	}
	return c.JSON(model.WebResponse{
		Success: true,
		Message: "profil berhasil diambil",
		Data:    u,
	})
}

// CheckPassword membandingkan password asli dengan hash bcrypt di database.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
