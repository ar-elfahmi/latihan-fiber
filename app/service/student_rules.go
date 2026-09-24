package service

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx,
// tidak menyentuh database, dan tidak tahu apa-apa tentang HTTP.
// Semua fungsi di sini bisa diuji hanya dengan `go test`.

import (
	"latihan-fiber/app/model"
)

// sortColumns whitelist untuk ORDER BY tabel students.
// Dipakai oleh student_service.go untuk validasi awal sort.
var sortColumns = map[string]bool{
	"id": true, "nim": true, "name": true, "grade": true, "created_at": true,
}

// IsEmptyPatch menandai PATCH yang tidak mengubah apa-apa.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada.
// Field pointer nil dibiarkan; field dikirim menimpa nilai lama.
// Validasi nilai (wajib, format, rentang) dilakukan oleh tag validate
// pada struct request; ApplyPatch murni menyalin.
func ApplyPatch(current model.Student, req model.PatchStudentRequest) model.Student {
	if req.NIM != nil {
		current.NIM = *req.NIM
	}
	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

// CountTotalPages membulatkan ke atas tanpa bilangan pecahan.
// Legacy: dipakai oleh List dengan OFFSET (sebelum D.3 pakai cursor).
func CountTotalPages(total, limit int) int {
	if limit <= 0 || total <= 0 {
		return 0
	}
	if total%limit == 0 {
		return total / limit
	}
	return total/limit + 1
}
