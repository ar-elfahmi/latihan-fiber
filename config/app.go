package config

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/app/service"
	"latihan-fiber/route"
	"log/slog"
	"time"
)

func NewApp(logger *slog.Logger, h *service.StudentHandler, pool *pgxpool.Pool) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "Tugas Mandiri REST API Students"})
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		id := fmt.Sprintf("req-%d", time.Now().UnixNano())
		c.Set("X-Request-ID", id)
		e := c.Next()
		logger.Info("request", slog.String("request_id", id), slog.String("method", c.Method()), slog.String("path", c.Path()), slog.Int("status", c.Response().StatusCode()), slog.Int64("duration_ms", time.Since(start).Milliseconds()))
		return e
	})
	route.Register(app, h, pool)
	return app
}
