package middleware

import (
	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/service"
	"latihan-fiber/helper"
)

// RequirePermission adalah middleware otorisasi berbasis permission.
//
// Cara pakai:
//
//	api.Get("/students/:id",
//	    middleware.RequirePermission("student:read:any"),
//	    s.Student.Get)
//
// Keputusan dibaca dari c.Locals("perms") yang diisi oleh Protect.
// Bisa menerima daftar kandidat (group): handler memilih :any ATAU :self;
// helper HasAnyPermissionFromGroups mengecek semua.
//
// Policy:
//   - perms kosong / nil     -> 403 (fail closed; user terautentikasi tapi
//     tak punya hak teridentifikasi).
//   - salah satu cocok       -> lanjut.
//   - tidak ada yang cocok   -> 403.
func RequirePermission(candidates ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		perms := readPerms(c)
		// 401 hanya kalau TIDAK ada identitas (Protect sudah gagal duluan).
		// Kasus umum: token valid -> Protect men-set user_id & perms.
		if !service.HasAnyPermissionFromGroups(perms, candidates) {
			return helper.Error(c, 403, "tidak punya hak untuk aksi ini")
		}
		return c.Next()
	}
}

// readPerms membaca c.Locals("perms"); jika Protect gagal memenuhi isian
// (token sah Modul 5 tanpa klaim perms, dsb.) kembalikan map kosong -> 403.
func readPerms(c *fiber.Ctx) map[string]bool {
	v := c.Locals("perms")
	if v == nil {
		return map[string]bool{}
	}
	if m, ok := v.(map[string]bool); ok {
		return m
	}
	return map[string]bool{}
}
