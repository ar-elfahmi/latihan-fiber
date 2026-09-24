package service

// Aturan otorisasi DATA mahasiswa (ownership check).
// File ini MURNI: tidak mengimpor fiber, tidak mengimpor repository,
// tidak tahu apa-apa tentang HTTP. Hanya model + helper.
// Semua fungsi di sini bisa diuji langsung lewat `go test`.

import (
	"latihan-fiber/app/model"
	"latihan-fiber/helper"
)

// CanAccessStudent memutuskan boleh atau tidaknya `current` mengakses
// satu baris data mahasiswa yang dimilik oleh `ownerID`.
//
// Aturannya dua lapis, sesuai tugas C.2.3:
//  1. Pemegang anyPermission (mis. "student:read:any") boleh mengakses
//     data mahasiswa siapa pun, termasuk yang tanpa pemilik.
//  2. Pemilik data (current.ID == ownerID) selalu boleh, walau dia
//     hanya punya permission :self.
//  3. Selain dua itu -> tolak.
//
// Fail closed pada data yang tidak bisa dipercaya: current.ID <= 0 atau
// ownerID <= 0 tanpa anyPermission mengembalikan false. perms nil berarti
// jalur :any otomatis gagal; satu-satunya jalan tersisa adalah kecocokan
// kepemilikan (aturan "pemilik selalu boleh").
func CanAccessStudent(current model.AuthUser, ownerID int, perms *helper.PermissionSet, anyPermission string) bool {
	// Lapis 1: izin global (:any) menang untuk data siapa pun.
	if perms.Has(anyPermission) {
		return true
	}

	// Lapis 2: kecocokan kepemilikan.
	// ID tidak valid dianggap gagal, bukan dianggap cocok.
	if current.ID <= 0 || ownerID <= 0 {
		return false
	}
	return current.ID == ownerID
}
