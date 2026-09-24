package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger mencatat setiap request: method, path, status, user_id, role,
// dan durasi. Dipasang sekali di main.go lewat app.Use(middleware.Logger).
//
// user_id dan role dibaca SETELAH c.Next() karena Logger dipasang paling
// awal (sebelum Protect). Saat c.Next() selesai, Protect sudah menaruh
// identitas di c.Locals. Request tanpa token (mis. login) tercatat
// user_id=0 role=- (pola Langkah 8 modul).
func Logger(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()

	userID, _ := c.Locals("user_id").(int)
	roles, _ := c.Locals("roles").([]string)
	if len(roles) == 0 {
		roles = []string{"-"}
	}

	log.Printf("%s %s user_id=%d role=%s status=%d %v",
		c.Method(), c.Path(), userID, strings.Join(roles, ","), c.Response().StatusCode(), time.Since(start))
	return err
}
