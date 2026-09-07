package service

import (
	"strings"

	"latihan-fiber/app/model"
)

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx,
// tidak menyentuh database, dan tidak tahu apa pun tentang HTTP.
// Seluruh fungsi di sini dapat diuji hanya dengan memanggilnya.

// ValidateCreate memeriksa isi permintaan pembuatan mahasiswa (POST).
// Mengembalikan peta berisi field yang bermasalah; kosong berarti lolos.
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "harus bernilai antara 0.0 - 4.0"
	}

	return errs
}

// ValidateUpdate memeriksa isi permintaan PUT.
// Seluruh field wajib ada karena PUT mengganti isi secara keseluruhan.
func ValidateUpdate(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 4 {
		errs["grade"] = "harus bernilai antara 0.0 - 4.0"
	}

	return errs
}

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada.
// Field yang bernilai nil dibiarkan apa adanya.
func ApplyPatch(
	current model.Student, req model.PatchStudentRequest,
) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM != nil {
		if v := strings.TrimSpace(*req.NIM); v == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			current.NIM = v
		}
	}

	if req.Name != nil {
		if v := strings.TrimSpace(*req.Name); v == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = v
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 4 {
			errs["grade"] = "harus bernilai antara 0.0 - 4.0"
		} else {
			current.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

// IsEmptyPatch menandai permintaan PATCH yang tidak mengubah apa pun.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil &&
		req.Grade == nil && req.IsActive == nil
}

// CountTotalPages membulatkan ke atas tanpa memakai bilangan pecahan.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
