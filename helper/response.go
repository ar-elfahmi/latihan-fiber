package helper

import "github.com/gofiber/fiber/v2"

// Error mengirim respons gagal dengan status dan pesan yang dipilih.
// Pesan jadi JSON: {"error": "..."}.
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}
