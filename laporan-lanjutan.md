# Lanjutan Laporan Tugas 3

> Catatan untuk saya sendiri: teks dalam kurung siku seperti [Gambar 1] adalah tempat
> untuk menempel screenshot dari Postman. Nomornya sudah sesuai urutan screenshot yang
> saya ambil. Bagian ini disambung setelah bagian penyesuaian handler, main, dan helper
> di laporan Word.

## Penjelasan Keunikan NIM

Kolom nim di tabel students diberi constraint UNIQUE. Artinya database akan menolak
INSERT atau UPDATE yang membuat dua baris punya NIM yang sama.

Aturan seperti ini sebenarnya bisa saya taruh di kode Go. Caranya, sebelum menyimpan
data, jalankan SELECT dulu untuk mengecek apakah NIM sudah dipakai. Kalau belum ada,
baru jalankan INSERT. Tapi cara ini punya dua kelemahan.

Kelemahan pertama adalah race condition. Dua request bisa datang hampir bersamaan.
Keduanya menjalankan SELECT, keduanya mendapat jawaban bahwa NIM belum dipakai, lalu
keduanya menjalankan INSERT. Hasilnya NIM jadi ganda, padahal kodenya sudah mengecek
terlebih dahulu. Constraint UNIQUE tidak punya celah seperti ini karena pengecekan
dan penyimpanan dilakukan database dalam satu langkah, tidak bisa disisip request lain
di antaranya.

Kelemahan kedua, kode Go bukan satu-satunya jalan masuk data. Data bisa ditambah lewat
psql, lewat aplikasi lain, atau saat restore backup. Kalau aturannya hanya ada di kode
Go, jalan masuk lain tetap bisa membuat NIM ganda. Dengan UNIQUE di database, semua
jalan masuk otomatis dijaga.

Jadi validasi di kode Go tetap berguna supaya kesalahan ketahuan lebih cepat, tapi
penjaga terakhirnya ada di database.

## Penjelasan Indeks Tambahan

Selain primary key, saya menambahkan satu indeks:

```sql
CREATE INDEX students_name_idx ON students (LOWER(name));
```

Indeks ini dipakai untuk pencarian nama. Endpoint GET /api/v1/students punya parameter
search yang di repository diterjemahkan menjadi `WHERE name ILIKE '%...%'`. ILIKE itu
pencarian yang tidak membedakan huruf besar dan kecil.

Tanpa indeks, PostgreSQL harus membaca seluruh isi tabel baris per baris baru bisa
tahu data mana yang cocok. Ibaratnya mencari kata di buku tanpa daftar isi, harus
membuka halaman satu per satu. Dengan indeks, database punya jalan pintas ke baris yang
maksudnya saja. Karena pencariannya memakai ILIKE, indeksnya saya buat di LOWER(name)
supaya formatnya cocok dengan cara pencariannya.

Jujur saja, di tabel yang isinya baru beberapa baris bedanya belum terasa. Manfaatnya
baru terlihat kalau datanya sudah ribuan. Untuk melihat efeknya bisa menjalankan
EXPLAIN ANALYZE di psql dan membandingkan rencana eksekusinya.

## Pengujian dengan Postman

Sebelum pengujian, saya buat dulu collection di Postman. Semua request memakai variabel
{{baseUrl}} yang isinya http://localhost:3000, jadi kalau nanti alamat servernya ganti,
cukup ubah satu tempat. Request saya bagi ke tiga folder: 00 Health, 01 CRUD Students,
dan 02 Bukti Error yang isinya skenario error khusus untuk laporan ini.

Setiap request di folder Bukti Error saya beri script assertion. Isinya sederhana,
misalnya memastikan responsnya harus 404. Jadi status code tidak hanya dilihat manual,
tetapi dicek otomatis oleh Postman, dan hasilnya muncul di panel Test Results di bawah
respons. Panel itu ikut terlihat di screenshot sebagai bukti bahwa tesnya lulus.

Ringkasan sepuluh screenshot pengujian:

| No | Request | Status |
|----|---------|--------|
| 1 | GET /health saat database hidup | 200 |
| 2 | GET /students dengan query page, limit, sort, order | 200 |
| 3 | GET /students/999 | 404 |
| 4 | PUT /students/999 | 404 |
| 5 | PATCH /students/999 | 404 |
| 6 | DELETE /students/999 | 404 |
| 7 | POST /students dengan NIM yang sudah dipakai | 409 |
| 8 | PUT /students/3 dengan NIM milik mahasiswa lain | 409 |
| 9 | GET /health setelah PostgreSQL dimatikan | 503 |
| 10 | GET /students setelah PostgreSQL dimatikan | 500 |

### Bukti Status 404

Untuk menguji 404 saya memakai id 999 yang pasti tidak ada di tabel. Ketika data tidak
ditemukan, repository mengirim error ErrNotFound, lalu handler menerjemahkannya menjadi
respons 404 dengan pesan mahasiswa tidak ditemukan.

[Gambar 3: GET id 999 menghasilkan 404]

[Gambar 4: PUT id 999 menghasilkan 404]

[Gambar 5: PATCH id 999 menghasilkan 404]

[Gambar 6: DELETE id 999 menghasilkan 404]

### Bukti Status 409

Untuk 409 lewat POST, saya mengirim NIM 434241047 yang sudah dipakai mahasiswa lain.
Untuk 409 lewat PUT, saya mengubah mahasiswa id 3 supaya memakai NIM milik mahasiswa
id 1. Database menolak keduanya dengan kode error 23505 (unique_violation), repository
menerjemahkannya menjadi ErrDuplicate, lalu handler membalas dengan 409.

[Gambar 7: POST dengan NIM ganda menghasilkan 409]

[Gambar 8: PUT dengan NIM milik mahasiswa lain menghasilkan 409]

### Bukti Saat Database Dimatikan

Setelah itu saya mematikan PostgreSQL dari Laragon, lalu mengulang dua request pertama.
Server Go masih hidup, jadi permintaan tetap sampai ke API, tetapi setiap operasi ke
database gagal.

[Gambar 9: GET /health saat database mati menghasilkan 503]

[Gambar 10: GET /students saat database mati menghasilkan 500]

## 503 atau 500, Mana yang Lebih Tepat

Dua status di atas muncul dari keadaan yang sama, yaitu database mati, tetapi kodenya
berbeda. Endpoint /health membalas 503, sedangkan endpoint data membalas 500.

500 artinya ada kesalahan tak terduga di sisi server. Kesan yang diberikan adalah ada
yang salah dengan programnya, misalnya bug, dan tidak jelas kapan bisa normal. Klien
tidak bisa berbuat banyak selain melaporkan errornya.

503 artinya servernya hidup dan kode berjalan normal, tapi layanan sedang tidak bisa
melayani permintaan. Status ini biasa dipakai saat server sedang maintenance atau
dependensinya sedang down. Ada makna tersirat bahwa kondisinya sementara dan klien bisa
mencoba lagi nanti.

Menurut saya 503 lebih tepat untuk keadaan database mati. Penyebabnya bukan bug di
kode, servernya sendiri masih hidup dan bisa membalas, dan kondisinya cenderung
sementara karena tinggal menyalakan PostgreSQL lagi. Jadi menurut saya pesan yang paling
jujur untuk dikirim ke klien adalah 503, bukan 500.

Kondisi API saya sekarang belum sepenuhnya begitu. Baru /health yang membalas 503 saat
database mati. Endpoint data masih 500 karena semua error database yang tidak dikenal
ditangani sebagai internal server error. Sebagai ide perbaikan, handler bisa mengenali
error koneksi database dari pgx, lalu membalasnya dengan 503, supaya semua endpoint
konsisten.

## Kesimpulan

API mahasiswa sekarang menyimpan datanya di PostgreSQL, bukan lagi di memori. Perilaku
HTTP-nya tidak berubah, hanya sumber datanya yang berbeda.

Beberapa hal yang saya pelajari dari tugas ini. Pertama, aturan data seperti keunikan
NIM lebih baik dijaga database karena bebas dari race condition dan berlaku untuk semua
jalan masuk data. Kedua, penyaringan, pencarian, pengurutan, dan paginasi yang dulu
dikerjakan di Go sekarang dikerjakan SQL, dengan ILIKE untuk pencarian, WHERE dengan
parameter untuk penyaringan, ORDER BY dengan daftar kolom yang diizinkan, serta LIMIT
dan OFFSET untuk paginasi. Ketiga, error database perlu dipetakan ke status HTTP yang
benar, 404 untuk data yang tidak ada, 409 untuk NIM ganda, dan 503 atau 500 untuk
gangguan database, dan menurut saya 503 yang paling tepat untuk kondisi database mati.
