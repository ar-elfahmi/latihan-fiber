package route

import (
	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/middleware"
)

// Setup mendaftarkan seluruh endpoint aplikasi.
// Tiap endpoint satu baris; aturan/bisnis ada di service.
// Modul 6: protect + permission dipasang per method agar authorization
// tidak bergantung pada isi data (pemilikan dicek service).
func Setup(app *fiber.App, s service.Services, rbac repository.RBACRepository, jwtSecret []byte) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Login: terbitkan token JWT (tidak perlu token untuk masuk ke sini).
	api.Post("/auth/login", s.Auth.Login)

	// Semua route yang didaftarkan SETELAH baris ini wajib bawa token.
	// Modul 6: Protect juga memuat permission live dari RBACRepository.
	api.Use(middleware.Protect(rbac, jwtSecret))

	// --- Daftar prestasi (global) ---
	api.Get("/prestasi",
		middleware.RequirePermission("prestasi:list"),
		s.Prestasi.List,
	)

	// --- Mahasiswa ---
	st := api.Group("/students")
	st.Get("/", middleware.RequirePermission("student:list"), s.Student.List)
	st.Get("/:id",
		middleware.RequirePermission("student:read:any", "student:read:self"),
		s.Student.Get,
	)
	st.Post("/", middleware.RequirePermission("student:create"), s.Student.Create)
	st.Put("/:id",
		middleware.RequirePermission("student:update:any", "student:update:self"),
		s.Student.Replace,
	)
	st.Patch("/:id",
		middleware.RequirePermission("student:update:any", "student:update:self"),
		s.Student.Patch,
	)
	st.Delete("/:id", middleware.RequirePermission("student:delete"), s.Student.Delete)

	// --- Prestasi (nested di bawah mahasiswa) ---
	st.Get("/:id/prestasi",
		middleware.RequirePermission("prestasi:read:any", "prestasi:read:self"),
		s.Prestasi.GetByStudent,
	)
	st.Post("/:id/prestasi",
		middleware.RequirePermission("prestasi:create"),
		s.Prestasi.Create,
	)
}
