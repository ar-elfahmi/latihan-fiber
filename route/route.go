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

// Dependencies adalah kumpulan dependency yang dibutuhkan route.
// Dipasang satu struct supaya main.go tidak perlu mengingat urutan
// parameter dan supaya slice berikutnya (repository jadi interface)
// bisa menukar implementasi tanpa mengubah handler.
type Dependencies struct {
	Services    service.Services
	RBAC        repository.RBACRepository
	JWTSecret   []byte
	Pool        *pgxpool.Pool
	MongoClient *mongo.Client
}

// Setup mendaftarkan seluruh endpoint aplikasi.
// Tiap endpoint satu baris; aturan/bisnis ada di service.
// Modul 6: protect + permission dipasang per method agar authorization
// tidak bergantung pada isi data (pemilikan dicek service).
func Setup(app *fiber.App, deps Dependencies) {
	s := deps.Services
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", healthCheck(deps.Pool, deps.MongoClient))

	// Login: terbitkan token JWT (tidak perlu token untuk masuk ke sini).
	api.Post("/auth/login", s.Auth.Login)
	api.Post("/auth/register", s.Auth.Register)

	// Semua route yang didaftarkan SETELAH baris ini wajib bawa token.
	// Modul 6: Protect juga memuat permission live dari RBACRepository.
	api.Use(middleware.Protect(deps.RBAC, deps.JWTSecret))

	api.Get("/courses", s.Course.List)
	api.Post("/courses/:id/enroll", middleware.RequirePermission("course:enroll"), s.Course.Enroll)
	api.Delete("/courses/:id/enroll", middleware.RequirePermission("course:enroll"), s.Course.Drop)
	api.Post("/enrollments/swap", middleware.RequirePermission("course:enroll"), s.Course.Swap)
	api.Patch("/courses/:id", middleware.RequirePermission("course:manage"), s.Course.UpdateQuota)

	api.Get("/notifications/mine", s.Notification.Mine)
	api.Get("/admin/queue", middleware.RequirePermission("queue:monitor"), s.Notification.Monitor)

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

	// --- Portofolio mahasiswa (MongoDB) ---
	// Permission memakai permission mahasiswa yang sudah ada; kepemilikan
	// dicek lagi di service (student:read:self / student:update:self).
	st.Get("/:id/portfolio",
		middleware.RequirePermission("student:read:any", "student:read:self"),
		s.Portfolio.Get,
	)
	st.Put("/:id/portfolio",
		middleware.RequirePermission("student:update:any", "student:update:self"),
		s.Portfolio.Upsert,
	)
	st.Post("/:id/portfolio/projects",
		middleware.RequirePermission("student:update:any", "student:update:self"),
		s.Portfolio.AddProject,
	)
	st.Delete("/:id/portfolio/projects/:projectId",
		middleware.RequirePermission("student:update:any", "student:update:self"),
		s.Portfolio.DeleteProject,
	)

	// --- Stats teknologi (publik setelah login) ---
	api.Get("/portfolios/stats/technologies",
		middleware.RequirePermission("student:read:any", "student:read:self"),
		s.Portfolio.StatsTechnologies,
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
