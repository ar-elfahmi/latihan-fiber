package config

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/helper"
)

// ErrorHandler adalah Fiber error handler terpusat. Dipasang via:
//
//	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
//
// Tugasnya: menerjemahkan setiap error yang muncul dari handler atau
// service menjadi JSON seragam. Handler tidak perlu membentuk JSON
// sendiri; cukup return helper.BadRequest(...) atau helper.Internal(err).
//
// Cabang:
//
//	1. *AppError           -> pakai field langsung (status, code, message, fields)
//	2. *fiber.Error         -> terjemahkan status ke kode AppError yang stabil
//	3. error lain (unknown) -> helper.Internal (fail closed) + log
//
// Bentuk JSON:
//
//	{
//	  "success":    false,
//	  "code":       "VALIDATION_FAILED",
//	  "message":    "...",
//	  "fields":     {"name": "wajib diisi"},
//	  "request_id": "..."
//	}
func ErrorHandler(c *fiber.Ctx, err error) error {
	requestID, _ := c.Locals("request_id").(string)

	// 1. AppError -> langsung pakai field-nya.
	if ae := helper.AsAppError(err); ae != nil {
		body := fiber.Map{
			"success":    false,
			"code":       ae.Code,
			"message":    ae.Message,
			"request_id": requestID,
		}
		if len(ae.Fields) > 0 {
			body["fields"] = ae.Fields
		}
		if ae.Cause() != nil {
			log.Printf("ERROR [req=%s] %s: %v", requestID, ae.Code, ae.Cause())
		}
		return c.Status(ae.Status).JSON(body)
	}

	// 2. Fiber error -> terjemahkan status ke kode AppError.
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fiber.Map{
			"success":    false,
			"code":       fiberCodeToAppCode(fe.Code),
			"message":    fe.Message,
			"request_id": requestID,
		})
	}

	// 3. Default: fail closed. Error tak dikenal dibungkus sebagai 500.
	log.Printf("ERROR [req=%s] unhandled: %v", requestID, err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success":    false,
		"code":       helper.CodeInternal,
		"message":    "internal server error",
		"request_id": requestID,
	})
}

// fiberCodeToAppCode menerjemahkan HTTP status dari fiber error menjadi
// kode AppError yang stabil. Pemetaan disimpan di sini supaya ErrorHandler
// tetap satu-satunya tempat penerjemahan status fiber ke kode aplikasi.
func fiberCodeToAppCode(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return helper.CodeBadRequest
	case fiber.StatusUnauthorized:
		return helper.CodeUnauthorized
	case fiber.StatusForbidden:
		return helper.CodeForbidden
	case fiber.StatusNotFound:
		return helper.CodeNotFound
	case fiber.StatusMethodNotAllowed:
		return helper.CodeBadRequest
	case fiber.StatusNotAcceptable:
		return helper.CodeNotAcceptable
	case fiber.StatusConflict:
		return helper.CodeConflict
	case fiber.StatusRequestEntityTooLarge:
		return helper.CodeUnsupportedMediaType
	case fiber.StatusUnsupportedMediaType:
		return helper.CodeUnsupportedMediaType
	case fiber.StatusUnprocessableEntity:
		return helper.CodeValidationFailed
	case fiber.StatusTooManyRequests:
		return helper.CodeTooManyRequests
	case fiber.StatusServiceUnavailable:
		return helper.CodeServiceUnavailable
	default:
		return helper.CodeInternal
	}
}
