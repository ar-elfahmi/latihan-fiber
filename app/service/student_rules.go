package service

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx,
// tidak menyentuh database, dan tidak tahu apa-apa tentang HTTP.
// Semua fungsi di sini bisa diuji hanya dengan `go test`.

import (
	"strings"

	"latihan-fiber/app/model"
)

// sortColumns whitelist untuk ORDER BY tabel students.
// Dipakai oleh student_service.go untuk validasi awal sort.
var sortColumns = map[string]bool{
	"id": true, "nim": true, "name": true, "grade": true, "created_at": true,
}

// ValidateCreate memeriksa body POST /students.
// Mengembalikan peta field→pesan; peta kosong berarti lolos.
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "harus antara 0.0 dan 4.0"
	}
	return errs
}

// ValidateReplace memeriksa body PUT /students/:id.
// Seluruh field wajib ada karena PUT menggantikan data total.
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "harus antara 0.0 dan 4.0"
	}
	return errs
}

// IsEmptyPatch menandai PATCH yang tidak mengubah apa-apa.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada.
// Field nil dibiarkan apa adanya. Mengembalikan data baru + peta error.
func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}
	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		if nim == "" {
			errs["nim"] = "wajib diisi"
		} else {
			current.NIM = nim
		}
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			errs["name"] = "wajib diisi"
		} else {
			current.Name = name
		}
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 4 {
			errs["grade"] = "harus antara 0.0 dan 4.0"
		} else {
			current.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current, errs
}

// CountTotalPages membulatkan ke atas tanpa bilangan pecahan.
func CountTotalPages(total, limit int) int {
	if limit <= 0 || total <= 0 {
		return 0
	}
	if total%limit == 0 {
		return total / limit
	}
	return total/limit + 1
}
