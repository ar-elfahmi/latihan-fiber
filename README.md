# latihan-fiber

API REST data mahasiswa + prestasi, dibangun dengan **Go + Fiber v2 + PostgreSQL (pgx v5)**.

## Filosofi Kode

> **Kesederhanaan adalah kunci.** Semua kode ditulis seramah mungkin supaya bisa dikembangkan sendiri tanpa AI.

- Tidak pakai dependency luar yang tidak perlu (tidak ada godotenv, lumberjack, slog)
- Pakai library standar Go (`log`, `os`, `context`)
- Database: `pgx.Connect` (koneksi tunggal, cocok belajar; production pakai pgxpool)
- Struct konkret saja, **tidak ada interface** repository
- Handler, validasi, dan SQL diletakkan di tempat yang paling logis

## Teknologi

- Go 1.26+
- Fiber v2 — web framework
- PostgreSQL — database
- pgx v5 — driver PostgreSQL

## Struktur Proyek

```
main.go              entry point: koneksi DB → repo → service → Fiber
config/logger.go     var Logger = log.Default()
config/env.go        os.Getenv + GetWithDefault
database/postgres.go pgx.Connect
middleware/middleware.go  fungsi Logger() saja
helper/request.go    ParseBody(c, target)
helper/response.go   Error(c, status, msg)
route/route.go       Setup(app, Services{Student, Prestasi})
app/model/           Student, Prestasi, request structs
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
4. **Salin `.env.example` ke `.env`** lalu isi password:
   ```env
   APP_PORT=3000
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=yourpassword
   DB_NAME=praktikum_backend
   DB_SSLMODE=disable
   ```
5. **Jalankan**:
   ```bash
   go run .
   ```
6. **Cek health**: `GET http://localhost:3000/api/v1/health`

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

- **Single connection (`pgx.Conn`)** — tidak thread-safe. Jika beban request tinggi, ganti `database.Connect` pakai `pgxpool.New` (sedikit perubahan saja).
- **Graceful shutdown** — tidak dipakai (Fiber `Listen` blocking sudah cukup untuk belajar).
- **Test** — `app/service/*_test.go` menguji business rules tanpa database/server.