package route

import (
	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/service"
	"latihan-fiber/middleware"
)

// Setup mendaftarkan seluruh endpoint aplikasi.
// Tiap endpoint satu baris; aturan/bisnis ada di service.
func Setup(app *fiber.App, s service.Services, jwtSecret []byte) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})


	// Login: terbitkan token JWT (tidak perlu token untuk masuk ke sini).
	api.Post("/auth/login", s.Auth.Login)

	// Semua route yang didaftarkan SETELAH baris ini wajib bawa token.
	api.Use(middleware.Protect(jwtSecret))
	// Mahasiswa
	st := api.Group("/students")
	st.Get("/", s.Student.List)
	st.Get("/:id", s.Student.Get)
	st.Post("/", s.Student.Create)
	st.Put("/:id", s.Student.Replace)
	st.Patch("/:id", s.Student.Patch)
	st.Delete("/:id", s.Student.Delete)

	// Prestasi (nested di bawah mahasiswa)
	st.Get("/:id/prestasi", s.Prestasi.GetByStudent)
	st.Post("/:id/prestasi", s.Prestasi.Create)
}
