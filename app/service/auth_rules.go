package service

import (
	"strings"

	"latihan-fiber/app/model"
)

// ValidateLogin memeriksa body POST /auth/login.
// Sama seperti rules lain: murni Go, tanpa fiber/database, bisa go test.
func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}
	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}
	return errs
}
