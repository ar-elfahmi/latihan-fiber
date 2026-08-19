package main

import "fmt"

func main() {

	nama := "Budi"      // string  → teks
	umur := 20          // int     → bilangan bulat
	ipk := 3.85         // float64 → bilangan desimal
	aktif := true       // bool    → true / false
	matkul := []string{ // slice   → seperti array, tapi ukurannya fleksibel
		"Matematika",
		"Pemrograman",
		"Basis Data",
	}

	// Tampilkan semua variabel
	fmt.Println("DATA MAHASISWA:")
	fmt.Println("Nama   :", nama)
	fmt.Println("Umur   :", umur)
	fmt.Println("IPK    :", ipk)
	fmt.Println("Aktif  :", aktif)
	fmt.Println("Matkul :", matkul)

	//map pasangan key-value
	nilaiMahasiswa := map[string]float64{
		"Budi":  3.85,
		"Ani":   3.50,
		"Citra": 3.70,
	}

	//tambah data baru ke map
	nilaiMahasiswa["Dodi"] = 3.20
	fmt.Println("\n[TAMBAH] Dodi ditambahkan ke map")

	// bisa true atau false, tergantung apakah key ada di map atau tidak
	nilai, ada := nilaiMahasiswa["Ani"]
	if ada {
		fmt.Println("[BACA]  Nilai Ani:", nilai)
	} else {
		fmt.Println("[BACA]  Ani tidak ditemukan")
	}

	// Coba baca nama yang TIDAK ada
	nilai2, ada2 := nilaiMahasiswa["Eko"]
	if ada2 {
		fmt.Println("[BACA]  Nilai Eko:", nilai2)
	} else {
		fmt.Println("[BACA]  Eko tidak ditemukan di map")
	}

	// Hapus data dari map
	delete(nilaiMahasiswa, "Citra")
	fmt.Println("[HAPUS] Citra dihapus dari map")

	// Tampilkan semua data mahasiswa
	fmt.Println("\n[TELUSURI] Semua data mahasiswa:")
	for namaMhs, nilaiMhs := range nilaiMahasiswa {
		fmt.Printf("  %-10s → %.2f\n", namaMhs, nilaiMhs)
	}
}
