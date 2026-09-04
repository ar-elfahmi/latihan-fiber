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

func Register(app *fiber.App, h *service.StudentHandler, pool *pgxpool.Pool) {
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if e := pool.Ping(ctx); e != nil {
			return helper.Fail(c, 503, "database tidak dapat dihubungi")
		}
		return helper.OK(c, "server dan database berjalan", nil)
	})
	g := app.Group("/api/v1/students", middleware.RequireJSON)
	g.Get("/", h.List)
	g.Get("/:id", h.Get)
	g.Post("/", h.Create)
	g.Put("/:id", h.Replace)
	g.Patch("/:id", h.Patch)
	g.Delete("/:id", h.Delete)
	app.Use(func(c *fiber.Ctx) error { return helper.Fail(c, 404, "endpoint tidak ditemukan") })
}
