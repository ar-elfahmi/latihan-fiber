Tugas Mandiri 
Pindahkan API mahasiswa (students) yang Anda buat pada tugas pertemuan 2 dari memori ke PostgreSQL. 
Seluruh perilaku HTTP-nya tidak boleh berubah; yang berubah hanya tempat datanya disimpan. 
1. Skema dan Migrasi 
• Buat berkas migrations/001_create_students.sql berisi tabel students 
• Kolom minimal: id, nim, name, grade, is_active, created_at, dengan tipe dan batasan yang Anda pilih 
sendiri 
• NIM wajib unik. Jelaskan di laporan mengapa keunikan itu lebih baik dijaga basis data daripada oleh 
kode Go 
• Tambahkan minimal satu indeks selain kunci primer, dan jelaskan alasannya 
2. Konfigurasi dan Koneksi 
• Simpan seluruh kredensial di .env, dan pastikan berkas itu tidak ikut ter-commit 
• Sertakan .env.example berisi nama variabel dengan nilai kosong 
• Buat database/postgres.go yang membangun connection pool dan melakukan Ping 
• Endpoint /health harus ikut memeriksa kondisi basis data 
3. Repository 
• Buat interface StudentRepository berisi minimal lima method: FindAll, FindByID, Create, Update, dan 
Delete 
• Buat implementasinya untuk PostgreSQL pada app/repository/student_repository.go 
• Seluruh nilai dari klien wajib dikirim sebagai parameter query, bukan disambung ke teks SQL 
• Sediakan sentinel error ErrNotFound dan ErrDuplicate 
• Tidak boleh ada satu pun penyebutan fiber di dalam paket repository 
4. Query di Sisi Basis Data 
Penyaringan, pencarian, pengurutan, dan paginasi yang pada pertemuan 2 dikerjakan di dalam Go harus 
dipindahkan seluruhnya ke SQL. 
Kemampuan 
Harus memakai 
Pencarian pada nama 
ILIKE 
Penyaringan berdasarkan nilai field 
Klausa WHERE dengan parameter 
Pengurutan 
ORDER BY dengan daftar putih kolom 
Paginasi 
LIMIT dan OFFSET 
meta.total 
SELECT COUNT(*) dengan syarat penyaringan yang sama 
5. Status HTTP dari Error Basis Data 
Modul Pertemuan 3 — Database & Repository Pattern 
Buktikan dengan tangkapan layar bahwa API Anda mengembalikan: 
Status 
Situasi yang harus Anda buat 
404 
GET, PUT, PATCH, atau DELETE pada id yang tidak ada di tabel 
409 
Menambah atau mengubah data sehingga NIM menjadi ganda 
503 atau 500 
Layanan PostgreSQL dimatikan, lalu API dipanggil 
Untuk status terakhir, jelaskan di laporan status mana yang menurut Anda paling tepat dan mengapa. Tidak 
ada satu jawaban benar; yang dinilai adalah alasannya. 
Sertakan pula di README.md: skema tabel Anda, cara menyiapkan basis data dari nol, dan daftar variabel 
environment yang diperlukan. Anggap pembacanya adalah rekan sekelas yang baru mengklona repositori 
Anda dan tidak bisa bertanya kepada Anda.