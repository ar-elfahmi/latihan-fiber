package middleware

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/helper"
)

// Logger mencatat setiap request: method, path, status, user_id, role,
// dan durasi. Dipasang sekali di main.go lewat app.Use(middleware.Logger).
//
// Status dibaca dari variabel lokal yang diisi SEBELUM c.Next() dan
// kemudian disesuaikan bila c.Next() mengembalikan error. Pola ini
// memperbaiki bug Modul 7 #5: variabel status yang sudah disimpan
// sempat diabaikan dan diganti pembacaan ulang setelah c.Next() yang
// nilainya sudah ter-reset.
//
// user_id dan role dibaca SETELAH c.Next() karena Logger dipasang paling
// awal (sebelum Protect). Saat c.Next() selesai, Protect sudah menaruh
// identitas di c.Locals. Request tanpa token (mis. login) tercatat
// user_id=0 role=-.
func Logger(c *fiber.Ctx) error {
	start := time.Now()
	status := c.Response().StatusCode() // variabel lokal, dipakai saat log
	err := c.Next()
	if err != nil {
		status = fiberErrorStatus(err)
	}

	userID, _ := c.Locals("user_id").(int)
	roles, _ := c.Locals("roles").([]string)
	if len(roles) == 0 {
		roles = []string{"-"}
	}

	log.Printf("%s %s user_id=%d role=%s status=%d %v",
		c.Method(), c.Path(), userID, strings.Join(roles, ","), status, time.Since(start))
	return err
}

// fiberErrorStatus mengambil HTTP status dari error. Dipakai Logger
// untuk menyesuaikan status lokal saat c.Next() mengembalikan error.
// Urutan prioritas: AppError, fiber.Error, fallback 500.
func fiberErrorStatus(err error) int {
	if ae := helper.AsAppError(err); ae != nil {
		return ae.Status
	}
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code
	}
	return fiber.StatusInternalServerError
}
