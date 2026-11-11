package config

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/helper"
)

// decodeBody membaca body respons sebagai map JSON agar kunci yang
// muncul (dan yang tidak muncul) bisa diperiksa.
func decodeBody(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("body bukan JSON valid: %v (body=%s)", err, string(body))
	}
	return out
}

// TestErrorHandlerKeepsErrorContract mengunci bentuk JSON error:
// {success, code, message, fields, request_id} (D-002).
func TestErrorHandlerKeepsErrorContract(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/uji", func(c *fiber.Ctx) error {
		return helper.Validation(map[string]string{"nama": "wajib diisi"})
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/uji", nil))
	if err != nil {
		t.Fatalf("request uji gagal: %v", err)
	}
	if resp.StatusCode != 422 {
		t.Errorf("status harus 422, dapat %d", resp.StatusCode)
	}

	body := decodeBody(t, readAll(t, resp))
	if body["success"] != false {
		t.Errorf("success harus false, dapat %v", body["success"])
	}
	if body["code"] != helper.CodeValidationFailed {
		t.Errorf("code harus %s, dapat %v", helper.CodeValidationFailed, body["code"])
	}
	if body["message"] != "validasi gagal" {
		t.Errorf("message harus 'validasi gagal', dapat %v", body["message"])
	}
	fields, ok := body["fields"].(map[string]any)
	if !ok || fields["nama"] != "wajib diisi" {
		t.Errorf("fields harus berisi nama, dapat %v", body["fields"])
	}
	if _, ok := body["request_id"]; !ok {
		t.Error("kunci request_id harus selalu ada")
	}
}

// TestErrorHandlerOmitsEmptyFields memastikan "fields" tidak muncul bila
// kosong — sama seperti perilaku sebelum refactor.
func TestErrorHandlerOmitsEmptyFields(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/uji", func(c *fiber.Ctx) error {
		return helper.NotFound("mahasiswa")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/uji", nil))
	if err != nil {
		t.Fatalf("request uji gagal: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("status harus 404, dapat %d", resp.StatusCode)
	}

	body := decodeBody(t, readAll(t, resp))
	if _, ok := body["fields"]; ok {
		t.Error("fields harus tidak muncul saat kosong")
	}
	if body["code"] != helper.CodeNotFound {
		t.Errorf("code harus %s, dapat %v", helper.CodeNotFound, body["code"])
	}
}

// TestErrorHandlerUnknownErrorFailsClosed memastikan error tak dikenal
// jadi 500 dengan kode INTERNAL.
func TestErrorHandlerUnknownErrorFailsClosed(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/uji", func(c *fiber.Ctx) error {
		return errors.New("kegagalan tak terduga")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/uji", nil))
	if err != nil {
		t.Fatalf("request uji gagal: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("status harus 500, dapat %d", resp.StatusCode)
	}

	body := decodeBody(t, readAll(t, resp))
	if body["code"] != helper.CodeInternal {
		t.Errorf("code harus %s, dapat %v", helper.CodeInternal, body["code"])
	}
	if body["message"] != "internal server error" {
		t.Errorf("message harus generik, dapat %v", body["message"])
	}
}

// readAll membaca seluruh body respons.
func readAll(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("gagal membaca body: %v", err)
	}
	return data
}
