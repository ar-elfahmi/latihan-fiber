package service

import "latihan-fiber/app/model"

// ValidatePrestasi memeriksa body POST /students/:id/prestasi.
// Aturan: keterangan wajib (maks 100 karakter); juara 0-100.
func ValidatePrestasi(req model.CreatePrestasiRequest) map[string]string {
	errs := map[string]string{}
	if len([]rune(req.Keterangan)) == 0 || len([]rune(req.Keterangan)) > 100 {
		errs["keterangan"] = "keterangan harus diisi (maksimal 100 karakter)"
	}
	if req.Juara < 0 || req.Juara > 100 {
		errs["juara"] = "juara harus berada di antara 0 dan 100"
	}
	return errs
}
