package helper

import (
	"errors"
	"fmt"
	"net/http"
)

// Konstanta kode error. Kode ini stabil dan boleh diandalkan oleh
// klien (mis. frontend switch berdasarkan Code, bukan pada pesan).
const (
	CodeBadRequest           = "BAD_REQUEST"
	CodeValidationFailed     = "VALIDATION_FAILED"
	CodeUnauthorized         = "UNAUTHORIZED"
	CodeForbidden            = "FORBIDDEN"
	CodeNotFound             = "NOT_FOUND"
	CodeConflict             = "CONFLICT"
	CodeNotAcceptable        = "NOT_ACCEPTABLE"
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	CodeTooManyRequests      = "TOO_MANY_REQUESTS"
	CodeInternal             = "INTERNAL"
	CodeServiceUnavailable   = "SERVICE_UNAVAILABLE"
	CodeInvalidCursor        = "INVALID_CURSOR"
)

// AppError adalah error aplikasi yang siap dikirim ke klien.
//
//	Status  : HTTP status code untuk response.
//	Code    : kode stabil untuk klien (UPPER_SNAKE_CASE).
//	Message : pesan aman untuk manusia, tanpa detail teknis.
//	Fields  : peta field ke pesan untuk validasi per-field.
//	cause   : error asli untuk log diagnostik. Tidak diserialkan.
//
// ErrorHandler terpusat (di config/app.go) yang mengubah *AppError
// menjadi JSON. Pemanggil cukup return helper.BadRequest(...) atau
// helper.Internal(err); handler tidak perlu menulis JSON sendiri.
type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	cause   error
}

// Error memenuhi interface error agar AppError bisa dicetak lewat log.
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Cause mengembalikan error asli yang dibungkus. Dipakai oleh
// middleware logger untuk mencatat akar masalah tanpa membocorkan
// detail teknis ke klien.
func (e *AppError) Cause() error {
	return e.cause
}

// Unwrap memenuhi interface errors.Unwrap agar errors.Is dan
// errors.As dapat menelusuri chain sampai ke error asli.
func (e *AppError) Unwrap() error {
	return e.cause
}

// AsAppError mengambil *AppError dari error chain. Mengembalikan
// nil bila err bukan AppError dan bukan pembungkusnya.
func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}

// Konstruktor untuk tiap kategori error. Semua menerima pesan
// singkat (atau nama resource untuk NotFound) dan mengembalikan
// *AppError siap pakai.

// BadRequest: input klien tidak valid secara umum.
// Pakai Validation() untuk kesalahan per-field.
func BadRequest(message string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: CodeBadRequest, Message: message}
}

// Validation: input klien tidak lolos aturan field-level.
// Status 422 Unprocessable Entity.
func Validation(fields map[string]string) *AppError {
	return &AppError{
		Status:  http.StatusUnprocessableEntity,
		Code:    CodeValidationFailed,
		Message: "validasi gagal",
		Fields:  fields,
	}
}

// Unauthorized: tidak ada token atau token tidak valid.
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "butuh autentikasi"
	}
	return &AppError{Status: http.StatusUnauthorized, Code: CodeUnauthorized, Message: message}
}

// Forbidden: terautentikasi tapi tidak berhak.
func Forbidden(message string) *AppError {
	if message == "" {
		message = "akses ditolak"
	}
	return &AppError{Status: http.StatusForbidden, Code: CodeForbidden, Message: message}
}

// NotFound: resource yang diminta tidak ada. Resource adalah nama
// benda (mis. "mahasiswa") yang disisipkan ke pesan standar.
func NotFound(resource string) *AppError {
	if resource == "" {
		resource = "sumber"
	}
	return &AppError{
		Status:  http.StatusNotFound,
		Code:    CodeNotFound,
		Message: resource + " tidak ditemukan",
	}
}

// Conflict: bentrok dengan data yang sudah ada (mis. NIM duplikat).
func Conflict(message string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: CodeConflict, Message: message}
}

// NotAcceptable: header Accept tidak dapat dipenuhi.
func NotAcceptable(message string) *AppError {
	return &AppError{Status: http.StatusNotAcceptable, Code: CodeNotAcceptable, Message: message}
}

// UnsupportedMediaType: Content-Type body tidak dikenal.
func UnsupportedMediaType(message string) *AppError {
	return &AppError{Status: http.StatusUnsupportedMediaType, Code: CodeUnsupportedMediaType, Message: message}
}

// TooManyRequests: rate limit (cadangan untuk pengembangan nanti).
func TooManyRequests(message string) *AppError {
	return &AppError{Status: http.StatusTooManyRequests, Code: CodeTooManyRequests, Message: message}
}

// ServiceUnavailable: layanan di bawah (mis. DB down).
func ServiceUnavailable(message string) *AppError {
	return &AppError{Status: http.StatusServiceUnavailable, Code: CodeServiceUnavailable, Message: message}
}

// Internal: error tak terduga. Pesan untuk klien dibuat generik;
// error asli dibungkus dan akan dicatat oleh middleware logger.
func Internal(cause error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternal,
		Message: "internal server error",
		cause:   cause,
	}
}

// InvalidCursor: cursor base64 tidak dapat di-decode.
func InvalidCursor() *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: CodeInvalidCursor, Message: "cursor tidak valid"}
}
