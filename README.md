# latihan-fiber

API REST data mahasiswa + prestasi, dibangun dengan **Go + Fiber v2 + PostgreSQL (pgx v5)**.

## Filosofi Kode

> **Kesederhanaan adalah kunci.** Semua kode ditulis seramah mungkin supaya bisa dikembangkan sendiri tanpa AI.

- Tidak pakai dependency luar yang tidak perlu (`go.mod` hanya: fiber, pgx, jwt, crypto)
- Pakai library standar Go (`log`, `os`, `context`)
- Database: `pgxpool` (connection pool, aman untuk request bersamaan)
- Struct konkret saja, **tidak ada interface** repository
- Handler, validasi, dan SQL diletakkan di tempat yang paling logis

## Teknologi

- Go 1.26+
- Fiber v2 — web framework
- PostgreSQL — database
- pgx v5 — driver PostgreSQL
- MongoDB — database portofolio mahasiswa (driver `go.mongodb.org/mongo-driver/v2`)

## Struktur Proyek

```
main.go              entry point: koneksi DB → repo → service → Fiber
config/env.go        LoadEnv / Get / GetWithDefault
database/postgres.go pgxpool
database/mongo.go    koneksi MongoDB (portofolio)
middleware/middleware.go  fungsi Logger(); middleware/auth.go  fungsi Protect()
helper/request.go    ParseBody(c, target)
helper/response.go   Error(c, status, msg)
route/route.go       Setup(app, Services{Student, Prestasi, User, Portfolio})
app/model/           Student, Prestasi, Portfolio, request structs
app/repository/      StudentRepository, PrestasiRepository (struct + metode SQL)
app/service/         StudentService, PrestasiService (struct + handler + rules)
migrations/          001_create_students.sql, 002_create_prestasi.sql
```

## Skema Database

### `students`
| Kolom | Tipe | Keterangan |
|-------|------|------------|
| id | SERIAL | PK, auto increment |
| nim | VARCHAR(50) | unik, wajib |
| name | VARCHAR(255) | wajib |
| grade | DOUBLE PRECISION | 0.0–4.0 (CHECK) |
| is_active | BOOLEAN | default TRUE |
| created_at | TIMESTAMPTZ | default NOW() |

### `prestasi`
| Kolom | Tipe | Keterangan |
|-------|------|------------|
| id | SERIAL | PK |
| student_id | INTEGER | FK → students(id), ON DELETE CASCADE |
| keterangan | VARCHAR(100) | wajib, maks 100 karakter |
| juara | INTEGER | 0–100 (CHECK) |
| created_at | TIMESTAMPTZ | default NOW() |

## Persiapan

1. **PostgreSQL jalan** (misalnya lewat Laragon)
2. **Buat database**:
   ```bash
   psql -U postgres -c "CREATE DATABASE praktikum_backend;"
   ```
3. **Jalankan migrasi**:
   ```bash
   psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
   psql -U postgres -d praktikum_backend -f migrations/002_create_prestasi.sql
   ```
4. **MongoDB jalan** (portofolio disimpan di MongoDB, bukan PostgreSQL). Koleksi `student_portfolios` dibuat otomatis saat server start.
5. **Salin `.env.example` ke `.env`** lalu isi password:
   ```env
   APP_PORT=3000
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=yourpassword
   DB_NAME=praktikum_backend
   DB_SSLMODE=disable
   MONGO_URI=mongodb://localhost:27017
   MONGO_DATABASE=praktikum_backend
   ```
6. **Jalankan**:
   ```bash
   go run .
   ```
7. **Cek health**: `GET http://localhost:3000/api/v1/health`

## Endpoint

| Method | Path | Keterangan |
|--------|------|------------|
| GET | /api/v1/health | cek server + DB |
| GET | /api/v1/students | daftar mahasiswa (paginasi, search, sort) |
| GET | /api/v1/students/:id | detail mahasiswa |
| POST | /api/v1/students | tambah mahasiswa |
| PUT | /api/v1/students/:id | ganti seluruh data |
| PATCH | /api/v1/students/:id | ubah sebagian data |
| DELETE | /api/v1/students/:id | hapus mahasiswa |
| GET | /api/v1/students/:id/prestasi | prestasi milik mahasiswa |
| POST | /api/v1/students/:id/prestasi | tambah prestasi |
| GET | /api/v1/students/:id/portfolio | portofolio mahasiswa (MongoDB) |
| PUT | /api/v1/students/:id/portfolio | simpan portofolio (headline, bio, certificates) |
| POST | /api/v1/students/:id/portfolio/projects | tambah proyek ke portofolio (maks 20) |
| DELETE | /api/v1/students/:id/portfolio/projects/:projectId | hapus proyek dari portofolio |
| GET | /api/v1/portfolios/stats/technologies | 10 teknologi paling banyak dipakai |

### Query param GET /students
| Param | Default | Contoh |
|-------|---------|--------|
| page | 1 | ?page=2 |
| limit | 10 (maks 50) | ?limit=5 |
| search | — | ?search=ahmad |
| sort | id | ?sort=name |
| order | asc | ?order=desc |
| is_active | — | ?is_active=true |

## Validasi

**Mahasiswa (POST/PUT)**:
- `nim` & `name` wajib diisi
- `grade` 0.0–4.0

**Prestasi (POST)**:
- `keterangan` wajib, maks 100 karakter
- `juara` 0–100

**Portofolio (PUT)**:
- `headline` wajib, maks 200 karakter
- `bio` wajib, maks 1000 karakter
- `certificates` opsional, maks 20 butir

**Proyek (POST)**:
- `title` wajib, maks 200 karakter
- `technologies` minimal 1, maksimal 10 (tiap item 2–50 karakter)
- `url` wajib, harus URL valid
- `year` 2000–2100
- jumlah proyek per portofolio maksimal 20

**PATCH**: field yang dikirim saja yang diubah; kirim body kosong → 400.

## Respons Error

Semua error pakai format JSON seragam:
```json
{ "error": "pesan kesalahan" }
```
Validasi gagal (422):
```json
{ "error": "validasi gagal", "fields": { "nim": "wajib diisi" } }
```

## Perintah Berguna

```bash
go run .          # jalankan server
go build ./...    # cek kompilasi
go vet ./...      # analisis statis
go test ./...     # jalankan unit test (business rules saja)
```

## Catatan Penting

- **Connection pool (`pgxpool`)** — `database.Connect` mengembalikan pool yang aman dipakai bersamaan oleh banyak request Fiber. Jangan diganti ke `pgx.Conn` tunggal: request yang bertabrakan akan error 500.
- **Graceful shutdown** — tidak dipakai (Fiber `Listen` blocking sudah cukup untuk belajar).
- **Test** — `app/service/*_test.go` menguji business rules tanpa database/server.
- **Otorisasi portofolio** — route portofolio memakai permission mahasiswa yang sudah ada (`student:read:any`/`self` untuk baca, `student:update:any`/`self` untuk tulis). Pemegang `:self` hanya boleh portofolio mahasiswa yang ia miliki (`owner_id`); portofolio tanpa pemilik hanya bisa diakses pemegang `:any`.