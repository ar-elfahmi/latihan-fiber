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

	// Limiter membatasi percobaan login gagal per IP (Modul 12: 5 kali
	// gagal per menit per IP -> 429). Bila nil, dipakai limiter bawaan.
	Limiter *helper.LoginLimiter
}

// Register mendaftarkan akun baru (publik, tanpa token).
//
// @Summary		Registrasi akun
// @Description	Mendaftarkan akun mahasiswa baru. Username harus unik.
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		model.RegisterRequest	true	"Data pendaftaran"
// @Success		201	{object}	model.RegisterResponse
// @Failure		400	{object}	model.ErrorResponse
// @Failure		409	{object}	model.ErrorResponse
// @Failure		422	{object}	model.ErrorResponse
// @Router			/auth/register [post]
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
	return c.Status(fiber.StatusCreated).JSON(model.RegisterResponse{ID: id, Username: req.Username, Email: req.Email})
}

// Login memeriksa username + password, lalu memberi token JWT.
// Alur: baca body -> validasi -> cari user -> cocokkan password -> buat token.
//
// @Summary		Login dan terbitkan token JWT
// @Description	Memeriksa username + password lalu menerbitkan access token dan refresh token. Percobaan login dibatasi 5 kali gagal per menit per IP.
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		model.LoginRequest	true	"Kredensial login"
// @Success		200	{object}	model.LoginResponse
// @Failure		400	{object}	model.ErrorResponse
// @Failure		401	{object}	model.ErrorResponse
// @Failure		422	{object}	model.ErrorResponse
// @Failure		429	{object}	model.ErrorResponse
// @Router			/auth/login [post]
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

	// 3. Rate limit: percobaan login dibatasi per IP (Modul 12). Bodies
	//    rusak / input tidak valid tidak dihitung sebagai kegagalan.
	//    rusak / input tidak valid tidak dihitung sebagai kegagalan.
	limiter := s.Limiter
	if limiter == nil {
		limiter = helper.DefaultLoginLimiter()
	}
	ip := c.IP()
	if !limiter.Allow(ip) {
		return helper.TooManyRequests("Terlalu banyak percobaan login")
	}

	// 4. Cari akun. Salah username ATAU salah password: pesannya sama,
	//    supaya orang lain tidak tahu field mana yang keliru.
	u, err := s.Repo.GetByUsername(c.Context(), req.Username)
	if err != nil || !CheckPassword(req.Password, u.PasswordHash) {
		limiter.RecordFailure(ip)
		return helper.Unauthorized("username atau password salah")
	}
	limiter.Reset(ip)

	// 5. Buat token dan kirimkan ke client.
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
	return c.JSON(model.LoginResponse{
		Token:        token,
		RefreshToken: refresh,
		User:         model.LoginUser{ID: u.ID, Username: u.Username},
	})
}

// Refresh menukar refresh token dengan pasangan token baru (rotation):
// token lama langsung dicabut, pemakaian ulang → 401.
//
// @Summary		Perbarui access token
// @Description	Menukar refresh token dengan pasangan token baru (rotation). Token lama langsung dicabut.
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		model.RefreshRequest	true	"Refresh token"
// @Success		200	{object}	model.WebResponse
// @Failure		400	{object}	model.ErrorResponse
// @Failure		401	{object}	model.ErrorResponse
// @Failure		422	{object}	model.ErrorResponse
// @Router			/auth/refresh [post]
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
