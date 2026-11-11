package service

import (
	"latihan-fiber/app/domain"
	"latihan-fiber/app/model"
)

// Aturan domain KRS. Implementasi tunggalnya ada di app/domain (package daun
// tanpa Fiber/pgx/mongo) supaya app/service dan app/repository memakai sumber
// kebenaran yang sama tanpa siklus import. File ini hanya mengekspos API
// yang dipakai modul Modul 12 (service.MaxCreditsPerSemester,
// service.WithinCreditLimit, service.CreditsAfterSwap, service.NormalizeTags).

// MaxCreditsPerSemester adalah batas SKS yang boleh diambil satu mahasiswa
// dalam satu semester.
const MaxCreditsPerSemester = domain.MaxCreditsPerSemester

// WithinCreditLimit memberi tahu apakah SKS yang sudah diambil (current)
// ditambah SKS tambahan (delta) masih di dalam batas MaxCreditsPerSemester.
//
// Batasnya tepat: current+delta == MaxCreditsPerSemester masih lolos,
// current+delta > MaxCreditsPerSemester ditolak. Memeriksa batas di dalam
// transaksi (app/repository) memakai aturan yang sama.
func WithinCreditLimit(current, delta int) bool {
	return domain.WithinCreditLimit(current, delta)
}

// CreditsAfterSwap menghitung total SKS setelah kelas asal (oldCourse)
// diganti kelas tujuan (newCourse):
// SKS sekarang - SKS kelas asal + SKS kelas tujuan.
func CreditsAfterSwap(currentCredits int, oldCourse, newCourse model.Course) int {
	return domain.CreditsAfterSwap(currentCredits, oldCourse, newCourse)
}

// NormalizeTags merapikan daftar tag: memangkas spasi di kedua ujung,
// mengubah huruf menjadi kecil, membuang tag kosong, menghapus duplikat, dan
// mempertahankan urutan kemunculan pertama.
//
// TODO(S-004): dipakai untuk tag ulasan kelas (MongoDB). Sampai slice itu
// dikerjakan, fungsi ini belum dipakai kode produksi.
func NormalizeTags(tags []string) []string {
	return domain.NormalizeTags(tags)
}
