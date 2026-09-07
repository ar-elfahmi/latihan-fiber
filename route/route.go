package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

// Register memetakan URL ke method pada service.
//
// Perhatikan isi file ini: tidak ada logika bisnis, tidak ada query,
// tidak ada validasi. Hanya daftar alamat dan siapa yang melayaninya.
func Register(app *fiber.App, pool *pgxpool.Pool, userService *service.StudentService) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool))

	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", userService.List)
	students.Get("/:id", userService.Get)
	students.Post("/", userService.Create)
	students.Put("/:id", userService.Replace)
	students.Patch("/:id", userService.Patch)
	students.Delete("/:id", userService.Delete)
}

// healthCheck melaporkan kondisi layanan beserta databasenya.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
