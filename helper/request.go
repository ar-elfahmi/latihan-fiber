package helper

import "github.com/gofiber/fiber/v2"

// ParseBody mendekode body JSON request ke struct target.
// Gagal? Bisa langsung dipakai: helper.ParseBody(c, &req)
func ParseBody(c *fiber.Ctx, target any) error {
	return c.BodyParser(target)
}
