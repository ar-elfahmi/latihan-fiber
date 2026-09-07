package route

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
	"time"
)

func Register(app *fiber.App, s *service.StudentService, pool *pgxpool.Pool) {
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if e := pool.Ping(ctx); e != nil {
			return helper.Fail(c, 503, "database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	})
	g := app.Group("/api/v1/students", middleware.RequireJSON)
	g.Get("/", s.List)
	g.Get("/:id", s.Get)
	g.Post("/", s.Create)
	g.Put("/:id", s.Replace)
	g.Patch("/:id", s.Patch)
	g.Delete("/:id", s.Delete)
	app.Use(func(c *fiber.Ctx) error { return helper.Fail(c, 404, "endpoint tidak ditemukan") })
}
