package helper

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// validate dibuat SEKALI untuk seluruh aplikasi.
//
// validator.New() melakukan refleksi dan menyimpan hasilnya dalam cache
// internal. Membuatnya ulang pada setiap request berarti membuang cache
// tersebut berkali-kali tanpa alasan.
var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	// Tanpa ini, pesan error menyebut nama field Go ("Username"),
	// padahal client mengirim dan membaca nama JSON ("username").
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	// Aturan buatan sendiri. Aturan yang tidak disediakan library tetap
	// ditulis secara deklaratif sebagai tag, bukan dikembalikan menjadi
	// pemeriksaan manual yang tersebar di dalam service.
	_ = v.RegisterValidation("nospace", func(fl validator.FieldLevel) bool {
		return !strings.ContainsAny(fl.Field().String(), " \t\n\r")
	})

	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		for _, r := range fl.Field().String() {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
				r != '.' && r != '_' {
				return false
			}
		}
		return true
	})

	// nimformat: NIM harus 8-12 digit angka. Dipakai entity students
	// (Modul 7 Tugas Mandiri D.2).
	_ = v.RegisterValidation("nimformat", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if len(s) < 8 || len(s) > 12 {
			return false
		}
		for _, r := range s {
			if r < '0' || r > '9' {
				return false
			}
		}
		return true
	})

	// passwordStrength mengembalikan "" untuk password kuat, dan pesan
	// alasan untuk password lemah. Tag strongpassword valid saat
	// passwordStrength kosong (== "").
	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return passwordStrength(fl.Field().String()) == ""
	})

	return v
}

// ValidateStruct menjalankan seluruh aturan pada tag struct dan
// mengembalikan peta nama field ke pesan berbahasa Indonesia.
// Mengembalikan nil berarti tidak ada pelanggaran.
func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	// Terjadi bila yang dikirim bukan struct — itu kesalahan programmer,
	// bukan kesalahan pemakai API. Jangan diam-diam dianggap valid.
	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string]string{"": "objek yang divalidasi tidak sah"}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"": "validasi gagal"}
	}
	result := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		// Pakai comma-ok untuk periksa keberadaan key, bukan
		// `!exists` pada string (Modul 7 bug #3).
		if _, ok := result[fe.Field()]; !ok {
			result[fe.Field()] = messageFor(fe)
		}
	}
	return result
}

// messageFor menerjemahkan nama tag menjadi kalimat yang dapat dibaca
// pemakai. Daftar ini terpusat: menambah satu tag baru cukup menambah
// satu case di sini, tidak menyebar ke banyak file.
func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if fe.Kind() == reflect.String {
			return "minimal " + fe.Param() + " karakter"
		}
		return "nilai minimal " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return "maksimal " + fe.Param() + " karakter"
		}
		return "nilai maksimal " + fe.Param()
	case "alphanum":
		return "hanya boleh berisi huruf dan angka"
	case "nospace":
		return "tidak boleh mengandung spasi"
	case "username":
		return "hanya boleh huruf, angka, titik, dan garis bawah"
	case "nimformat":
		return "nim harus 8-12 digit angka"
	case "strongpassword":
		// Type assertion memakai bentuk DUA nilai, bukan satu. Bentuk
		// satu nilai akan panic bila suatu saat tag ini terpasang pada
		// field bukan string — mematikan server hanya karena salah tag.
		if value, ok := fe.Value().(string); ok {
			return passwordStrength(value)
		}
		return "password tidak memenuhi syarat"
	case "oneof":
		return "harus salah satu dari: " +
			strings.ReplaceAll(fe.Param(), " ", ",")
	default:
		// Jaring pengaman. Bila muncul di log, artinya ada tag yang
		// dipakai tetapi belum diterjemahkan di sini.
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

// passwordStrength mengembalikan:
//
//	"" untuk password yang kuat (lolos semua aturan);
//	pesan alasan untuk password yang lemah.
//
// Aturan: minimal 8 karakter, mengandung huruf besar, huruf kecil,
// dan angka.
func passwordStrength(p string) string {
	if len(p) < 8 {
		return "password minimal 8 karakter"
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	switch {
	case !hasUpper:
		return "password harus mengandung huruf besar"
	case !hasLower:
		return "password harus mengandung huruf kecil"
	case !hasDigit:
		return "password harus mengandung angka"
	}
	return ""
}
