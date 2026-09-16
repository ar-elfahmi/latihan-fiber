package model

// Prestasi adalah baris pada tabel prestasi (prestasi milik seorang mahasiswa).
type Prestasi struct {
	ID         int    `json:"id"`
	StudentID  int    `json:"student_id"`
	Keterangan string `json:"keterangan"`
	Juara      int    `json:"juara"`
}

// CreatePrestasiRequest adalah body POST /students/:id/prestasi.
// student_id diisi dari parameter rute (:id), bukan dari body,
// sehingga client tidak bisa menentukan sendiri kepemilikan data.
type CreatePrestasiRequest struct {
	Keterangan string `json:"keterangan"`
	Juara      int    `json:"juara"`
}
