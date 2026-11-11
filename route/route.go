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

	// Katalog permission per aksi (helper.NewPermissionSet). Nama
	// permission tidak ditulis inline supaya satu sumber kebenaran.
	perms := helper.NewPermissionSet()

	// Health check
	api.Get("/health", healthCheck(deps.Pool, deps.MongoClient))

	// Login: terbitkan token JWT (tidak perlu token untuk masuk ke sini).
	api.Post("/auth/login", s.Auth.Login)
	api.Post("/auth/register", s.Auth.Register)

	// Semua route yang didaftarkan SETELAH baris ini wajib bawa token.
	// Modul 6: Protect juga memuat permission live dari RBACRepository.
	api.Use(middleware.Protect(deps.RBAC, deps.JWTSecret))

	api.Get("/courses", s.Course.List)
	api.Post("/courses/:id/enroll", middleware.RequirePermission(perms.CourseEnroll...), s.Course.Enroll)
	api.Delete("/courses/:id/enroll", middleware.RequirePermission(perms.CourseEnroll...), s.Course.Drop)
	api.Post("/enrollments/swap", middleware.RequirePermission(perms.CourseEnroll...), s.Course.Swap)
	api.Patch("/courses/:id", middleware.RequirePermission(perms.CourseManage...), s.Course.UpdateQuota)

	api.Get("/notifications/mine", s.Notification.Mine)
	api.Get("/admin/queue", middleware.RequirePermission(perms.QueueMonitor...), s.Notification.Monitor)

	// --- Daftar prestasi (global) ---
	api.Get("/prestasi",
		middleware.RequirePermission(perms.PrestasiList...),
	)

	// --- Mahasiswa ---
	st := api.Group("/students")
	st.Get("/", middleware.RequirePermission(perms.StudentList...), s.Student.List)
	st.Get("/:id",
		middleware.RequirePermission(perms.StudentRead...),
		s.Student.Get,
	)
	st.Post("/", middleware.RequirePermission(perms.StudentCreate...), s.Student.Create)
	st.Put("/:id",
		middleware.RequirePermission(perms.StudentUpdate...),
		s.Student.Replace,
	)
	st.Patch("/:id",
		middleware.RequirePermission(perms.StudentUpdate...),
		s.Student.Patch,
	)
	st.Delete("/:id", middleware.RequirePermission(perms.StudentDelete...), s.Student.Delete)

	// --- Prestasi (nested di bawah mahasiswa) ---
	st.Get("/:id/prestasi",
		middleware.RequirePermission(perms.PrestasiRead...),
		s.Prestasi.GetByStudent,
	)
	st.Post("/:id/prestasi",
		middleware.RequirePermission(perms.PrestasiCreate...),
		s.Prestasi.Create,
	)

	// --- User management (Modul 7) ---
	api.Get("/users",
		middleware.RequirePermission(perms.UserList...),
		s.User.List,
	)

	// --- Portofolio mahasiswa (MongoDB) ---
	// Permission memakai permission mahasiswa yang sudah ada; kepemilikan
	// dicek lagi di service (student:read:self / student:update:self).
	st.Get("/:id/portfolio",
		middleware.RequirePermission(perms.StudentRead...),
		s.Portfolio.Get,
	)
	st.Put("/:id/portfolio",
		middleware.RequirePermission(perms.StudentUpdate...),
		s.Portfolio.Upsert,
	)
	st.Post("/:id/portfolio/projects",
		middleware.RequirePermission(perms.StudentUpdate...),
		s.Portfolio.AddProject,
	)
	st.Delete("/:id/portfolio/projects/:projectId",
		middleware.RequirePermission(perms.StudentUpdate...),
		s.Portfolio.DeleteProject,
	)

	// --- Stats teknologi (publik setelah login) ---
	api.Get("/portfolios/stats/technologies",
		middleware.RequirePermission(perms.StudentRead...),
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
