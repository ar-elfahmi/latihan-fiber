package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger mencatat method + path + durasi setiap request.
// Dipasang sekali di main.go lewat app.Use(middleware.Logger).
func Logger(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	log.Printf("%s %s %s %v", c.Method(), c.Path(), c.IP(), time.Since(start))
	return err
}
