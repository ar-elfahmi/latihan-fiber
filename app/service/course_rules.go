package service

import (
	"strings"

	"latihan-fiber/app/model"
)

// Aturan domain KRS. File ini adalah SATU-SATUNYA tempat implementasi aturan
// ini hidup: service memakainya langsung, dan app/repository menerimanya lewat
// injeksi (repository.CreditRules, wiring di main.go) supaya repository tidak
// mengimpor app/service (siklus import).
//
// API di sini juga yang dipakai modul Modul 12: service.MaxCreditsPerSemester,
// service.WithinCreditLimit, service.CreditsAfterSwap, service.NormalizeTags.

// MaxCreditsPerSemester adalah batas SKS yang boleh diambil satu mahasiswa
// dalam satu semester. Angkanya ada di app/model (sumber kebenaran tunggal)
// dan diekspos ulang di sini supaya API modul tetap hidup di package service.
const MaxCreditsPerSemester = model.MaxCreditsPerSemester

// WithinCreditLimit memberi tahu apakah SKS yang sudah diambil (current)
// ditambah SKS tambahan (delta) masih di dalam batas MaxCreditsPerSemester.
//
// Batasnya tepat: current+delta == MaxCreditsPerSemester masih lolos,
// current+delta > MaxCreditsPerSemester ditolak. delta boleh negatif
// (mis. saat melepas kelas).
func WithinCreditLimit(current, delta int) bool {
	return current+delta <= MaxCreditsPerSemester
}

// CreditsAfterSwap menghitung total SKS setelah kelas asal (oldCourse)
// diganti kelas tujuan (newCourse):
// SKS sekarang - SKS kelas asal + SKS kelas tujuan.
func CreditsAfterSwap(currentCredits int, oldCourse, newCourse model.Course) int {
	return currentCredits - oldCourse.Credits + newCourse.Credits
}

// NormalizeTags merapikan daftar tag: memangkas spasi di kedua ujung,
// mengubah huruf menjadi kecil, membuang tag kosong, menghapus duplikat, dan
// mempertahankan urutan kemunculan pertama.
//
// TODO(S-004): dipakai untuk tag ulasan kelas (MongoDB). Sampai slice itu
// dikerjakan, fungsi ini belum dipakai kode produksi.
func NormalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		clean := strings.ToLower(strings.TrimSpace(tag))
		if clean == "" {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		normalized = append(normalized, clean)
	}
	return normalized
}
