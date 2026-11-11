// Package domain berisi aturan domain murni yang dipakai lebih dari satu
// lapisan: app/service (API) dan app/repository (transaksi database).
//
// Package ini adalah daun dependensi: tidak mengimpor Fiber, pgx, mongo,
// maupun app/service. Karena itu aturannya bisa diuji tanpa database dan
// tanpa siklus import.
package domain

import (
	"strings"

	"latihan-fiber/app/model"
)

// MaxCreditsPerSemester adalah batas SKS yang boleh diambil satu mahasiswa
// dalam satu semester. Ini satu-satunya tempat angka batas SKS ditulis di
// seluruh kode produksi.
const MaxCreditsPerSemester = 24

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
