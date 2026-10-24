package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

// Setup mendaftarkan seluruh endpoint aplikasi.
// Tiap endpoint satu baris; aturan/bisnis ada di service.
// Modul 6: protect + permission dipasang per method agar authorization
// tidak bergantung pada isi data (pemilikan dicek service).
func Setup(app *fiber.App, s service.Services, rbac repository.RBACRepository, jwtSecret []byte, pool *pgxpool.Pool, mongoClient *mongo.Client) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", healthCheck(pool, mongoClient))

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

	// --- User management (Modul 7) ---
	api.Get("/users",
		middleware.RequirePermission("user:list"),
		s.User.List,
	)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}
		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
		}

		// Pakai c.JSON seperti kode lain di repo ini
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "server, PostgreSQL, dan MongoDB berjalan",
		})
	}
}
