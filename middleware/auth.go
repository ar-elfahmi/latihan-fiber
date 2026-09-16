package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/service"
	"latihan-fiber/helper"
)

// Protect menjaga endpoint: request wajib membawa token JWT yang sah.
// Caranya: tulis di header — Authorization: Bearer <token>
// Token diambil dari hasil login (POST /api/v1/auth/login).
func Protect(secret []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil header. Harus diawali kata "Bearer ".
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return helper.Error(c, 401, "token belum dikirim. Login dulu, lalu kirim header Authorization: Bearer <token>")
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")

		// 2. Periksa tanda tangan + kedaluwarsa.
		claims, err := service.ParseToken(tokenString, secret)
		if err != nil {
			return helper.Error(c, 401, "token tidak valid atau sudah kedaluwarsa")
		}

		// 3. Simpan identitas user agar handler berikutnya bisa memakai:
		//    userID := c.Locals("user_id").(float64)
		c.Locals("user_id", claims["sub"])
		c.Locals("username", claims["username"])
		return c.Next()
	}
}
