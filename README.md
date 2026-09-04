# latihan-fiber

Nama: Alfian Rasyid El Fahmi
NIM: 434241047

tugas ini saya kerjakan sendiri, beberapa hal seperti penjelasan konsep-konsep saya minta untuk di jelaskan ulang oleh claude.
sumber youtube yang saya tonton untuk pembelajaran sebagai berikut:
1. https://www.youtube.com/watch?v=85xEdlKojL4
2. https://youtu.be/4qH-7w5LZsA?si=DhmWNFj8dBN4IEsO
3. https://youtu.be/tC49Nzm6SyM?si=PYtgqcWW53IqoxXZ
4. https://youtu.be/446E-r0rXHI?si=KztHIE_SZvyOmwFd
5. https://youtu.be/rK87DPmXKss?si=6YksfVHOAzDyAIo6
6. https://youtu.be/FWGsVgFJWF4?si=Aj_VQFEZsDRpoMHY

## Tentang Proyek Ini

API REST untuk data mahasiswa yang dibuat dengan Go Fiber. Pada tugas pertemuan 2
datanya masih disimpan di memori. Sekarang penyimpanannya pindah ke PostgreSQL,
sedangkan perilaku HTTP-nya tidak berubah.

Teknologi yang dipakai:

- Go 1.26
- Fiber v2 sebagai framework web
- PostgreSQL sebagai database
- pgx v5 sebagai driver PostgreSQL
- godotenv untuk membaca berkas .env

## Struktur Proyek

```
api-students/        entry point API dan composition root
app/model/           entities: Student, request/response, dan ListQuery
app/service/         use case, business rules, dan HTTP handler
app/repository/      kontrak dan implementasi PostgreSQL
config/              perakitan Fiber, environment, dan logger JSON
database/            pembuatan connection pool PostgreSQL
helper/              presenter response dan pembaca request
middleware/          middleware global RequireJSON
route/               pendaftaran endpoint tanpa business rules
migrations/          berkas SQL untuk pembuatan tabel
tugas 1/             file latihan pertemuan 1, tidak ikut proses build
```

## Clean Architecture

Restrukturisasi pertemuan 4 memisahkan entity, business rules, adapter HTTP,
repository, konfigurasi, middleware, dan route. Dependency kode diarahkan ke
dalam: route memanggil service, service memakai interface repository, dan
implementasi PostgreSQL berada di luar. Business rules pada
`app/service/student_rules.go` tidak mengimpor Fiber sehingga dapat diuji tanpa
menyalakan server atau database. Laporan dan screenshot praktik tersedia di
`docs/Tugas4_434241047.md`.

## Skema Tabel

Tabel dibuat lewat `migrations/001_create_students.sql` dengan isi sebagai berikut:

```sql
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    grade DOUBLE PRECISION NOT NULL CHECK (grade >= 0.0 AND grade <= 4.0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS students_name_idx
ON students (LOWER(name));
```

Penjelasan singkat tiap kolom:

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | SERIAL | primary key, nilainya bertambah otomatis |
| nim | VARCHAR(50) | nomor induk mahasiswa, wajib unik |
| name | VARCHAR(255) | nama mahasiswa, wajib diisi |
| grade | DOUBLE PRECISION | nilai IPK, dibatasi 0.0 sampai 4.0 lewat CHECK |
| is_active | BOOLEAN | status aktif, default TRUE |
| created_at | TIMESTAMPTZ | waktu data dibuat, diisi otomatis NOW() |

Indeks tambahan `students_name_idx` dibuat pada `LOWER(name)` karena pencarian nama
memakai `ILIKE` yang tidak membedakan huruf besar dan kecil, jadi format indeksnya
harus cocok dengan cara pencariannya.

## Variabel Environment

Semua konfigurasi disimpan di berkas `.env` (tidak ikut ter-commit). Salin dari
`.env.example` lalu isi sesuai komputer masing-masing:

| Variabel | Contoh | Keterangan |
|---|---|---|
| APP_PORT | 3000 | port tempat server API berjalan |
| DB_HOST | localhost | alamat server PostgreSQL |
| DB_PORT | 5432 | port PostgreSQL |
| DB_USER | postgres | user untuk koneksi |
| DB_PASSWORD | (isi password Anda) | password user PostgreSQL |
| DB_NAME | praktikum_backend | nama database yang dipakai |
| DB_SSLMODE | disable | mode SSL koneksi, disable untuk lokal |
| DB_MAX_CONNS | 10 | jumlah maksimal koneksi di pool |

## Cara Menyiapkan Database dari Nol

Prasyarat: Go 1.26 atau lebih baru, PostgreSQL yang sudah jalan (misalnya lewat
Laragon), dan Git.

1. Clone repositori dan masuk ke foldernya:

   ```
   git clone https://github.com/ar-elfahmi/latihan-fiber.git
   cd latihan-fiber
   ```

2. Unduh dependensinya:

   ```
   go mod download
   ```

3. Buat database baru (nama harus sama dengan DB_NAME di .env):

   ```
   psql -U postgres -c "CREATE DATABASE praktikum_backend;"
   ```

4. Jalankan migrasi untuk membuat tabel:

   ```
   psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
   ```

   Catatan untuk pengguna Laragon di Windows: jika perintah psql tidak dikenali,
   tambahkan dulu foldernya ke PATH, misalnya
   `C:\laragon\bin\postgresql\postgresql\bin`, atau gunakan terminal bawaan Laragon.

5. Salin `.env.example` menjadi `.env` lalu isi `DB_PASSWORD` dengan password
   PostgreSQL Anda. Sesuaikan variabel lain jika setup Anda berbeda.

6. Jalankan API-nya:

   ```
   go run ./api-students
   ```

7. Buka `http://localhost:3000/api/v1/health`. Kalau muncul pesan
   `server dan database berjalan`, berarti semuanya sudah siap.

## Endpoint

| Method | Path | Keterangan |
|---|---|---|
| GET | /api/v1/health | cek kondisi server dan database |
| GET | /api/v1/students | daftar mahasiswa, mendukung query param |
| GET | /api/v1/students/:id | detail satu mahasiswa |
| POST | /api/v1/students | tambah mahasiswa baru |
| PUT | /api/v1/students/:id | ganti seluruh data mahasiswa |
| PATCH | /api/v1/students/:id | ubah sebagian data mahasiswa |
| DELETE | /api/v1/students/:id | hapus mahasiswa |

Query param yang didukung GET /api/v1/students:

| Param | Contoh | Keterangan |
|---|---|---|
| page | 1 | nomor halaman |
| limit | 10 | jumlah data per halaman, maksimal 50 |
| search | alfian | cari nama, memakai ILIKE |
| sort | name | kolom pengurutan: id, nim, name, grade, created_at |
| order | asc | arah pengurutan: asc atau desc |
| is_active | true | filter status aktif |

## Pengujian dengan Postman

Di repositori ini ada berkas `API Students v1.postman_collection.json`. Import
berkas itu ke Postman, buat environment dengan variabel `baseUrl` berisi
`http://localhost:3000` dan `studentId`, lalu pilih environment tersebut. Collection
sudah berisi folder khusus untuk skenario error (404, 409, 503) lengkap dengan
assertion status code di tiap requestnya.
