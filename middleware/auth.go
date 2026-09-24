package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/helper"
)

// Protect menjaga endpoint: request wajib membawa token JWT yang sah.
// Caranya: tulis di header — Authorization: Bearer <token>
// Token diambil dari hasil login (POST /api/v1/auth/login).
//
// Modul 6: setelah token sah, Protect ambil permission live dari database.
// Token tidak membawa klaim perms -> sumber kebenaran tunggal adalah DB,
// sehingga perubahan role langsung berlaku walau token masih hidup.
// - query permission gagal -> 403 (identitas ada, hak tak dapat ditentukan)
// - token invalid -> 401 (tabel 401/403 modul konsisten)
func Protect(rbac repository.RBACRepository, secret []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil header. Harus diawali kata "Bearer ".
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return helper.Unauthorized("token belum dikirim. Login dulu, lalu kirim header Authorization: Bearer <token>")
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")

		// 2. Periksa tanda tangan + kedaluwarsa.
		claims, err := service.ParseToken(tokenString, secret)
		if err != nil {
			return helper.Unauthorized("token tidak valid atau sudah kedaluwarsa")
		}

		// 3. Ambil user_id dari klaim.
		userID, err := claimInt(claims, "sub")
		if err != nil {
			return helper.Unauthorized("token tidak berisi identitas yang sah")
		}

		// 4. Muat permission live dari DB. Token valid -> identitas ada;
		//    gagal query -> 403 (hak tak dapat ditentukan). Fail closed.
		perms, err := rbac.GetPermissionsByUserID(c.Context(), userID)
		if err != nil {
			return helper.Forbidden("gagal memuat hak akses user")
		}

		// 4b. Ambil daftar role user. Hanya untuk pencatatan log (Langkah 8),
		// bukan untuk keputusan akses (keputusan tetap dari permission),
		// jadi kegagalannya tidak menutup akses yang sah.
		roles, roleErr := rbac.GetRolesByUserID(c.Context(), userID)
		if roleErr != nil {
			roles = nil
		}

		// 5. Simpan identitas + role + perms untuk handler berikutnya.
		c.Locals("user_id", userID)
		c.Locals("username", claims["username"])
		c.Locals("roles", roles)
		c.Locals("perms", perms)
		return c.Next()
	}
}

// claimInt mengubah klaim JWT (number -> float64 / string) menjadi int.
func claimInt(claims map[string]any, key string) (int, error) {
	v, ok := claims[key]
	if !ok || v == nil {
		return 0, errMissingClaim
	}
	switch x := v.(type) {
	case float64:
		return int(x), nil
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case string:
		return strconv.Atoi(x)
	default:
		return 0, errInvalidClaim
	}
}

var (
	errMissingClaim = strError("klaim identitas tidak ditemukan dalam token")
	errInvalidClaim = strError("klaim identitas tidak dapat dibaca")
)

type strError string

func (e strError) Error() string { return string(e) }
