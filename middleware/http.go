package middleware

import (
	"github.com/gofiber/fiber/v2"
	"latihan-fiber/helper"
	"strings"
)

func RequireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodPatch {
		if !strings.HasPrefix(c.Get("Content-Type"), fiber.MIMEApplicationJSON) {
			return helper.Fail(c, 415, "Content-Type harus application/json")
		}
	}
	return c.Next()
}
