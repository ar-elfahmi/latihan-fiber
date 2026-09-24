package service

import (
	"github.com/gofiber/fiber/v2"
)

// CurrentUserID mengambil user_id yang diletakkan middleware.Protect ke
// c.Locals("user_id"). Return 0 bila Protect belum sempat jalan (request
// masuk lewat endpoint publik).
func CurrentUserID(c *fiber.Ctx) int {
	v := c.Locals("user_id")
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	}
	return 0
}

// CurrentPerms mengambil himpunan permission yang sudah dimuat oleh Protect.
func CurrentPerms(c *fiber.Ctx) map[string]bool {
	v := c.Locals("perms")
	if v == nil {
		return map[string]bool{}
	}
	if m, ok := v.(map[string]bool); ok {
		return m
	}
	return map[string]bool{}
}
