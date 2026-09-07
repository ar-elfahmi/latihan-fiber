//go:build ignore

package main

import "fmt"

// FUNCTION 1: swap — menukar dua angka
// Tanda * di depan int artinya: "kirim alamat memorinya, bukan nilainya"

func swap(a *int, b *int) {
	temp := *a // simpan dulu nilai di alamat a
	*a = *b    // isi alamat a dengan nilai dari alamat b
	*b = temp  // isi alamat b dengan nilai temp yang disimpan tadi
}

// FUNCTION 2: updateSlice — tambah item ke slice
// *[]string artinya: pointer ke slice of string agar bisa diubah langsung di sumbernya

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem) // tambahkan newItem ke slice aslinya
}

// FUNCTION PERCOBAAN: pass by VALUE (tanpa pointer)
// Fungsi ini menerima SALINAN angka, bukan angka aslinya, jadi perubahan tidak akan terjadi di luar function

func swapSalah(a int, b int) {
	temp := a
	a = b
	b = temp
	// perubahan terjadi, tapi hanya di dalam function ini
	// variabel asli di luar tidak ikut berubah
}

func main() {
	// PERBANDINGAN: Pass by Value vs Pass by Pointer

	fmt.Println("PERBANDINGAN: VALUE vs POINTER")
	x := 10
	y := 20

	// Percobaan 1 — pakai function yang SALAH (pass by value)
	fmt.Println("\n--- Percobaan 1: Pass by VALUE (tanpa pointer) ---")
	fmt.Printf("Sebelum swap: x = %d, y = %d\n", x, y)
	swapSalah(x, y) // kirim nilai saja, bukan alamatnya
	fmt.Printf("Sesudah swap: x = %d, y = %d\n", x, y)
	fmt.Println("→ Hasilnya TIDAK berubah.")

	// Reset nilai
	x = 10
	y = 20

	// Percobaan 2 — pakai function yang BENAR (pass by pointer)
	fmt.Println("\n--- Percobaan 2: Pass by POINTER (dengan &) ---")
	fmt.Printf("Sebelum swap: x = %d, y = %d\n", x, y)
	swap(&x, &y) // kirim ALAMAT memorinya pakai tanda &
	fmt.Printf("Sesudah swap: x = %d, y = %d\n", x, y)
	fmt.Println("→ Hasilnya BERUBAH! Karena diubah langsung di sumbernya.")

	fmt.Println("DEMO: updateSlice")
	matkul := []string{"Matematika", "Pemrograman"}
	fmt.Println("\nSlice awal    :", matkul)

	updateSlice(&matkul, "Basis Data")
	fmt.Println("Setelah tambah:", matkul)

	updateSlice(&matkul, "Jaringan")
	fmt.Println("Setelah tambah:", matkul)
}
