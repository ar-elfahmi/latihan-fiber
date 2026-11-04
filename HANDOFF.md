# HANDOFF — `latihan-fiber` ( handing over ke Claude )

Dokumen ini adalah peta proyek + aturan kerja untuk whoever (Claude) yang
mengambil alih pengerjaan tugas berikutnya. **Baca seluruh dokumen ini sebelum mengubah kode.**

> Repo: `C:/laragon/www/01-code-kuliah/backend-lanjut/latihan-fiber`
> Proyek: Praktikum Pemrograman Backend Lanjut — Universitas Airlangga (D4 Teknik Informatika)
> NIM mahasiswa: `434241047`
> Stack: **Go 1.26 + Fiber v2 + PostgreSQL (pgx v5) + MongoDB (mongo-driver/v2)**

---

## 1. Ringkasan Proyek

REST API data **mahasiswa + prestasi + portofolio**, dengan autentikasi JWT dan
**RBAC (role-based access control)**. Portofolio mahasiswa disimpan di
**MongoDB** (bukan PostgreSQL) sebagai Introducing-NoSQL bagian modul.

Filosofi codebase (dari README): *kesederhanaan adalah kunci* — kode ditulis
supaya bisa dikembangkan sendiri, tanpa dependency berlebihan, tanpa
abstraksi interface yang tidak perlu.

### Modul yang sudah dikerjakan (status saat handoff)

| Modul | Topik | Status |
|---|---|---|
| 1–2 | Go + Fiber, routing, JSON | Selesai |
| 3 | Database + Repository Pattern | Selesai |
| 4 | Clean Architecture | Selesai |
| 5 | Authentication (JWT, bcrypt) | Selesai |
| 6 | Authorization & RBAC | Selesai (`migrations/004`) |
| 7 | Advanced API Design (AppError terpusat, validasi deklaratif, cursor pagination, content negotiation) | Selesai (`migrations/005`, `006`) |
| 8–9 | (modul lanjutan; belum ada berkas tugas di repo) | — |
| 10 | Introducing NoSQL — portofolio di MongoDB | Selesai (kode ada, wiring lengkap) |
| Modul baru (akan menyusul) | **belum diberikan** | Menunggu instruksi user |

Berkas materi tugas ada di `daftar tugas/` (git-ignored). Dokumen teknis
lokal ada di `docs/`.

---

## 2. Aturan Main (WAJIB dipatuhi)

Sumber: `AGENTS.md` (repo) + `C:\Users\lenovo\.pi\agent\AGENTS.md` (global).

1. **Bahasa**: komentar, pesan error, dan penjelasan = **Bahasa Indonesia**.
2. **Log** pakai `log.Println/Printf/Fatal` — **bukan** `fmt.Println`, bukan `slog`.
3. **SQL selalu parameterized** (`$1, $2, ...`). `ORDER BY` wajib lewat **whitelist map**.
4. **Struct konkret, tanpa interface** untuk repository Postgres.
   (Catatan: `repository.PortfolioRepository` *adalah* interface — itu pengecualian
   yang disengaja karena Mongo; jangan diubah tanpa alasan.)
5. **Business rules** harus di file `*_rules.go` yang **tidak mengimpor Fiber /
   database** → bisa diuji `go test` tanpa server.
6. **Error keluar lewat `helper.*AppError`**, tidak pernah `c.JSON` manual di handler.
7. **Jangan commit** `.exe`, `logs/`, `.env`, `daftar tugas/`.
8. Urutan menambah fitur: **migration → model → repository → rules → service → wiring (route.go, main.go)**.
9. **Never commit kecuali diminta.** Pesan commit harus conventional:
   `feat(scope): ...`, `fix(scope): ...`, `refactor(scope): ...`, `chore: ...`, `docs: ...`, `test(scope): ...`
10. Perubahan minimal & terfokus; jangan refactor tak diminta.

---

## 3. Struktur Direktori & Tanggung Jawab

```
main.go                Perakitan: LoadEnv → DB → Mongo → repo → service → Fiber
config/
  env.go               LoadEnv / Get / GetWithDefault (tanpa godotenv)
  app.go               ErrorHandler terpusat (AppError → JSON seragam)
  logger.go            var Logger = log.Default()
database/
  postgres.go          Connect() → *pgxpool.Pool (WAJIB pool, bukan pgx.Conn)
  mongo.go             NewMongoClient() → *mongo.Client (ping 5s, pool 20)
  test_setup.sql       Data uji
route/route.go         Setup(app, Services, rbacRepo, jwtSecret, pool, mongoClient)
middleware/
  middleware.go        Logger() — access log + durasi
  auth.go              Protect(rbac, secret) — validasi JWT + muat perms live
  authz.go             RequirePermission(candidates...)
helper/
  errors.go            AppError + konstruktor (BadRequest, Validation, Unauthorized,
                       Forbidden, NotFound, Conflict, NotAcceptable, InvalidCursor, Internal...)
  request.go           ParseBody(c, target)
  validator.go         validator global + tag custom + pesan Bahasa Indonesia
  cursor.go            EncodeCursor/DecodeCursor (base64 JSON {t,i})
  negotiate.go         Negotiate / ChosenFormat / WriteCSV
  permission.go        PermissionSet (map[string]bool, fail closed)
app/model/             Entity + request/response struct (student, prestasi, user,
                       portfolio, rbac, cursor)
app/repository/        SQL/Mongo konkret (student, prestasi, user, rbac, portfolio)
app/service/           Handler + business rules (student, prestasi, auth, user,
                       portfolio) + file *_rules.go (murni) + *_test.go
migrations/            001…006 SQL (lihat §8)
docs/                  panduan memahami proyek, alur endpoint, alur main.go
daftar tugas/          materi modul (git-ignored)
ai-usage.md            deklarasi penggunaan AI per tugas (wajib diisi tiap tugas!)
test-over-20-newData.js  skrip mongosh untuk menguji batas 20 proyek
```

**Arah dependensi (tidak boleh dibalik):**
`main → route → service → repository → model/database` · `helper` & `config` dipakai semua.

---

## 4. Menjalankan Proyek

### Prasyarat
- PostgreSQL jalan (Laragon). Database: `praktikum_backend`
- MongoDB jalan (port 27017)
- Go 1.26+

### Konfigurasi (`.env` — sudah ada, jangan commit)

```env
APP_PORT=3000
JWT_SECRET=rahasia-latihan
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=...
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=...
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=praktikum_backend
```

> `config.LoadEnv()` membaca `.env` **tanpa menimpa** env yang sudah ada di sistem.
> Nilai fallback dipakai lewat `config.GetWithDefault` (mis. `JWT_SECRET` → `rahasia-latihan`,
> `MONGO_DATABASE` → `praktikum_backend`).
> ⚠️ `.env.example` **belum** memuat `MONGO_URI`/`MONGO_DATABASE` — tambahkan bila relevan.

### Migrasi (jalankan berurutan, idempotent)
```bash
psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
psql -U postgres -d praktikum_backend -f migrations/002_create_prestasi.sql
psql -U postgres -d praktikum_backend -f migrations/003_create_users.sql
psql -U postgres -d praktikum_backend -f migrations/004_student_permissions.sql
psql -U postgres -d praktikum_backend -f migrations/005_cursor_index_users.sql
psql -U postgres -d praktikum_backend -f migrations/006_students_cursor_index.sql
```

### Perintah harian
```bash
go run .        # jalankan server (:3000)
go build ./...  # cek kompilasi
go vet ./...    # analisis statis
go test ./...   # unit test business rules (tanpa DB/server)
```

### Akun demo (dibuat migration 004)
| username | password | role |
|---|---|---|
| admin | admin123 | admin |
| staff | staff123 | staff |
| user  | user123  | user |

Hash password dibuat PostgreSQL (`pgcrypto`, bcrypt cost 10). Sample data:
mahasiswa `2101010101` (Budi Santoso) dimiliki `user`; `2101010102` (Sari Dewi) dimiliki `admin`.

---

## 5. Arsitektur Runtime (alur satu request)

```
HTTP → middleware.Logger → route handler chain:
        Protect(rbac, secret)          ← setelah /auth/login; semua route berikutnya wajib token
          → RequirePermission(...)     ← cek permission (kasar, tidak bergantung data)
            → service.Handler          ← cek kepemilikan (halus, bergantung data)
              → repository (SQL/Mongo)
```

- `main.go` **tidak** memuat logika bisnis; hanya perakitan + listener.
- `route/route.go` **tidak** memuat logika; hanya daftar URL → handler + middleware.
- Handler service **tidak** menulis JSON error manual; cukup `return helper.Xxx(...)`.

### Yang dipasang di `main.go`
1. `config.LoadEnv()`
2. `database.Connect` (pgxpool) → `defer db.Close()`
3. `database.NewMongoClient` → `defer Disconnect` (timeout 5s)
4. `repository.NewPortfolioRepository(mongoDB)` + `EnsureIndexes` (index unik `student_id`)
5. Repo Postgres konkret: Student / Prestasi / User / RBAC
6. `service.Services{Student, Prestasi, Auth, User, Portfolio}`
7. `fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})`
8. `app.Use(middleware.Logger)`
9. `route.Setup(...)`
10. 404 catch-all → `helper.NotFound("endpoint")`

---

## 6. Model Error Terpusat (WAJIB dipatuhi)

`helper/errors.go` → `AppError{Status, Code, Message, Fields, cause}`.

| Konstruktor | HTTP | Code |
|---|---|---|
| `BadRequest(msg)` | 400 | `BAD_REQUEST` |
| `Validation(fields)` | **422** | `VALIDATION_FAILED` (+ `fields` map) |
| `Unauthorized(msg)` | 401 | `UNAUTHORIZED` |
| `Forbidden(msg)` | 403 | `FORBIDDEN` |
| `NotFound(resource)` | 404 | `NOT_FOUND` |
| `Conflict(msg)` | 409 | `CONFLICT` |
| `NotAcceptable(msg)` | 406 | `NOT_ACCEPTABLE` |
| `UnsupportedMediaType` | 415 | `UNSUPPORTED_MEDIA_TYPE` |
| `TooManyRequests` | 429 | `TOO_MANY_REQUESTS` |
| `ServiceUnavailable` | 503 | `SERVICE_UNAVAILABLE` |
| `InvalidCursor()` | 400 | `INVALID_CURSOR` |
| `Internal(cause)` | 500 | `INTERNAL` (pesan generik, cause hanya di log) |

Bentuk JSON (dihasilkan `config.ErrorHandler`):
```json
{
  "success": false,
  "code": "VALIDATION_FAILED",
  "message": "validasi gagal",
  "fields": { "nim": "wajib diisi" },
  "request_id": ""
}
```
> `request_id` **saat ini selalu kosong** — belum ada middleware yang mengesetnya (lihat §12).

**Penerjemahan error repository → HTTP** ada di `app/service/error_translate.go`
(`translateError(err, resource)`):
`ErrNotFound`/`ErrStudentNotFound`→404 · `ErrDuplicate`→409 · `ErrProjectLimit`→409 ·
PG `23505`→409 · `23503`→400 · `23502`→400 · `23514`→400 · default→**500 fail closed**.

---

## 7. Autentikasi & Otorisasi (Modul 6)

### Token
- `POST /api/v1/auth/login` → `{"token": "...", "user": {"id":1,"username":"admin"}}`
- JWT HS256, klaim: `sub` (user id), `username`, `exp` (**12 jam**).
- **Token TIDAK membawa permission.** Sumber kebenaran tunggal = database.
  Konsekuensi: perubahan role berlaku seketika; DB tak terbaca → **403 fail closed**.
- Password: bcrypt (`golang.org/x/crypto/bcrypt`). Pesan login salah disamakan
  untuk username maupun password (anti user-enumeration).

### Middleware
- `middleware.Protect(rbac, secret)`:
  - Header wajib `Authorization: Bearer <token>` → tanpa itu **401**.
  - Token salah/kedaluwarsa/algoritma bukan HMAC → **401**.
  - Query permission gagal → **403** (identitas ada, hak tak dapat ditentukan).
  - Sukses → `c.Locals`: `user_id` (int), `username`, `roles` ([]string), `perms` (map[string]bool).
- `middleware.RequirePermission("a:any", "a:self")`:
  - perms kosong/nil/tidak cocok → **403**. (401 hanya bila tak ada identitas.)

### Otorisasi data (ownership) — di layer service
`app/service/student_authz_rules.go` → `CanAccessStudent(current model.AuthUser, ownerID int, perms *helper.PermissionSet, anyPerm string) bool`:
1. Pemegang `:any` → boleh untuk data siapa pun (termasuk tanpa pemilik).
2. Pemilik (`current.ID == ownerID`) → **selalu boleh**, walau hanya punya `:self`.
3. Selain itu → tolak. `current.ID <= 0` atau `ownerID <= 0` (owner NULL) → tolak (fail closed).

Bantu lain: `HasPermission`, `HasAnyPermissionFromGroups`, `RequireOwnerOrAdmin`,
`IsAdminOnlyPermission` (semua di `app/service/rbac_rules.go`).

### Matriks permission (dari `migrations/004_student_permissions.sql`)
| Permission | admin | staff | user |
|---|---|---|---|
| `user:list`, `user:read:any`, `user:update:any`, `user:delete`, `role:assign` | ✅ | ❌ | ❌ |
| `student:list`, `student:read:any`, `student:create` | ✅ | ✅ | ❌ |
| `student:update:any`, `student:delete` | ✅ | ❌ | ❌ |
| `student:read:self`, `student:update:self` | ❌ | ❌ | ✅ |
| `prestasi:list`, `prestasi:read:any`, `prestasi:create` | ✅ | ✅ | ❌ |
| `prestasi:read:self` | ❌ | ❌ | ✅ |

> Tabel RBAC: `roles`, `permissions`, `role_permissions`, `user_roles` (many-to-many).
> `students.owner_id` → `users(id) ON DELETE SET NULL` (NULL = belum ada pemilik).
> Semua statement migration **idempotent** (`IF NOT EXISTS` / `ON CONFLICT DO NOTHING`).

---

## 8. Skema Database

### PostgreSQL (`praktikum_backend`)
**`students`** (001 + 004 + 006)
`id SERIAL PK` · `nim VARCHAR(50) UNIQUE NOT NULL` · `name VARCHAR(255) NOT NULL` ·
`grade DOUBLE PRECISION NOT NULL CHECK (0..4)` · `is_active BOOLEAN NOT NULL DEFAULT TRUE` ·
`owner_id INT NULL REFERENCES users(id) ON DELETE SET NULL` · `created_at TIMESTAMPTZ DEFAULT NOW()`
Index: `students_name_idx (LOWER(name))`, `students_owner_id_idx`, index cursor id DESC (006).

**`prestasi`** (002)
`id SERIAL PK` · `student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE` ·
`keterangan VARCHAR(100) NOT NULL` · `juara INT NOT NULL CHECK (0..100)` · `created_at`.

**`users`** (003 + 005)
`id SERIAL PK` · `username TEXT UNIQUE` · `password_hash TEXT` · `created_at`.
Index cursor: `users_created_at_id_desc_idx (created_at DESC, id DESC)` — **arah DESC harus
cocok dengan `ORDER BY` di repository**, kalau terbalik planner pilih Seq Scan.

**RBAC** (004): `roles(name PK, description)` · `permissions(name PK, description)` ·
`role_permissions(role_name, permission_name)` · `user_roles(user_id, role_name)`.

### MongoDB (`praktikum_backend` → collection `student_portfolios`)
```jsonc
{
  "_id": ObjectId,
  "student_id": int,            // unique index "student_unique"
  "headline": string,
  "bio": string,
  "projects": [ { "_id": ObjectId, "title": String, "technologies": [String],
                  "url": String, "year": int } ],   // MAKS 20
  "certificates": [ { "name": String, "issuer": String, "date": String } ],
  "created_at": Date, "updated_at": Date
}
```
- Batas 20 proyek dijamin **atomik** di DB: filter `{"student_id": X, "projects.19": {"$exists": false}}` + `$push`.
- `StatsTechnologies` = aggregation pipeline `$unwind` → `$unwind` → `$group` → `$sort` → `$limit 10`.

---

## 9. Endpoint Lengkap

Prefix: `/api/v1`. Semua kecuali `/health` dan `/auth/login` **wajib** `Authorization: Bearer <token>`.

| Method | Path | Permission | Handler | Catatan |
|---|---|---|---|---|
| GET | `/health` | — (publik) | `route.healthCheck` | ping PG + Mongo, 503 bila salah satu mati |
| POST | `/auth/login` | — (publik) | `AuthService.Login` | body: `{username,password}` |
| GET | `/prestasi` | `prestasi:list` | `PrestasiService.List` | **masih offset pagination**, response `{data,page,limit,total,total_pages,sort,order}` |
| GET | `/students` | `student:list` | `StudentService.List` | **cursor pagination** + content negotiation (JSON/CSV) |
| GET | `/students/:id` | `student:read:any` \| `:self` | `StudentService.Get` | ownership check di service |
| POST | `/students` | `student:create` | `StudentService.Create` | `owner_id` **selalu** dari token (anti mass-assignment) |
| PUT | `/students/:id` | `student:update:any` \| `:self` | `StudentService.Replace` | ganti semua field |
| PATCH | `/students/:id` | `student:update:any` \| `:self` | `StudentService.Patch` | field pointer + `omitnil`; body kosong → 422 |
| DELETE | `/students/:id` | `student:delete` | `StudentService.Delete` | 204 tanpa body |
| GET | `/students/:id/prestasi` | `prestasi:read:any` \| `:self` | `PrestasiService.GetByStudent` | `student_id` dari path |
| POST | `/students/:id/prestasi` | `prestasi:create` | `PrestasiService.Create` | non-admin wajib student miliknya |
| GET | `/users` | `user:list` | `UserService.List` | cursor + CSV |
| GET | `/students/:id/portfolio` | `student:read:any` \| `:self` | `PortfolioService.Get` | Mongo |
| PUT | `/students/:id/portfolio` | `student:update:any` \| `:self` | `PortfolioService.Upsert` | headline/bio/certificates (maks 20) |
| POST | `/students/:id/portfolio/projects` | `student:update:any` \| `:self` | `PortfolioService.AddProject` | maks 20 proyek, 409 bila penuh |
| DELETE | `/students/:id/portfolio/projects/:projectId` | `student:update:any` \| `:self` | `PortfolioService.DeleteProject` | 404 bila proyek tidak ada |
| GET | `/portfolios/stats/technologies` | `student:read:any` \| `:self` | `PortfolioService.StatsTechnologies` | 10 teknologi terpopuler |

> `student:read:self` / `student:update:self` TIDAK ada di tabel role admin/staff —
> admin/staff memakai jalur `:any`.

---

## 10. Kontrak Request & Validasi (Modul 7)

Validasi **deklaratif** lewat tag `validate:` + `helper.ValidateStruct` (422 + map `fields`).
Validator global dibuat **sekali** (`helper/validator.go`), pesan memakai **nama JSON** field.

**Tag custom yang terdaftar** (tambah tag baru = tambah `case` di `messageFor`):
`nospace` · `username` (huruf/angka/./_) · `nimformat` (8–12 digit angka) ·
`strongpassword` (min 8, ada huruf besar + kecil + angka).

Request struct penting:
- `CreateStudentRequest` / `ReplaceStudentRequest`:
  `nim required,nimformat,max=20` · `name required,min=1,max=255` · `grade min=0,max=4` · `is_active`
- `PatchStudentRequest`: semua field pointer + `omitnil` → hanya field terkirim yang divalidasi/diterapkan (`ApplyPatch`).
- `CreatePrestasiRequest`: `keterangan` (manual rule `ValidatePrestasi`, maks 100), `juara` 0–100.
- `UpsertPortfolioRequest`: `headline required,max=200` · `bio required,max=1000` · `certificates omitempty,max=20,dive`.
- `AddProjectRequest`: `title required,max=200` · `technologies required,min=1,max=10,dive,min=2,max=50` · `url required,url` · `year required,min=2000,max=2100`.
- `CreateUserRequest` / `ReplaceUserRequest` / `PatchUserRequest` / `RegisterRequest` — **sudah ada di model, belum ada route/handler/repo** (lihat §12).

### Cursor pagination
- Cursor = base64url(JSON `{ "t": created_at RFC3339Nano, "i": id }`) → `helper.EncodeCursor/DecodeCursor`.
- Query param: `?cursor=<...>&limit=N` (default 10, maks 50). Cursor invalid → 400 `INVALID_CURSOR`.
- Pola: repository ambil `limit+1` baris → service potong jadi `limit` → `has_more = true` bila ada baris ekstra → `next_cursor` dari baris terakhir.
- **Tanpa** `total`/`total_pages` (see `model.CursorMeta`).
- `users`: klausa `(created_at, id) < ($1,$2)` ORDER BY `created_at DESC, id DESC`.
  `students`: `id < $1` ORDER BY `id DESC` (cursor time diabaikan).

### Content negotiation
`helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)` **dipanggil SEBELUM query DB**.
`Accept: application/json` (default/`*/*`) → JSON; `text/csv` → CSV; tidak cocok → **406**.
CSV ditulis via `encoding/csv` + `w.Flush()` **sebelum** `w.Error()` (bug Modul 7 #9).

---

## 11. Testing & Konvensi Git

- Test hanya di `app/service/*_rules_test.go` (rules murni). Currently:
  `TestHasPermission`, `TestHasAnyPermissionFromGroups`, `TestRequireOwnerOrAdmin`,
  `TestIsAdminOnlyPermission`, `TestCanAccessStudent`, `TestApplyPatch`,
  `TestIsEmptyPatch`, `TestCountTotalPages`.
- `go test ./...` tidak butuh DB/server.
- Test integrasi manual: Postman collection `API Students v1.postman_collection.json`;
  skrip mongosh `test-over-20-newData.js` (uji batas 20 proyek).
- `ai-usage.md` **wajib** diperbarui tiap tugas (kebijakan/praktik penggunaan AI).
- Commit message conventional; jangan commit `.env`, `.exe`, `logs/`, `daftar tugas/`.

---

## 12. Status quo: Gap & hal yang perlu diketahui (bukan untuk "dibuat-buat")

> Daftar ini penting agar AI tidak salah mengira fitur sudah ada.

1. **`request_id` kosong** — `config.ErrorHandler` membaca `c.Locals("request_id")`,
   tapi **tidak ada middleware yang mengesetnya**. Bila modul baru menuntut, tambahkan middleware
   request-id (generate UUID/ULID per request → set `c.Locals`) bila modul baru menuntut.
2. **CRUD user belum ada.** `UserService` hanya punya `List`. Struct
   `CreateUserRequest`, `ReplaceUserRequest`, `PatchUserRequest`, `RegisterRequest`
   sudah ada di `app/model/user.go`, tapi **tidak ada** method repository/service/route
   untuk create/update/delete user, dan **belum ada** `POST /auth/register`.
   Repository user juga belum punya `GetByID`, `Create`, `Replace`, `Delete`.
3. **`GET /students` kehilangan filter `search`/`is_active` dan `sort`.**
   Jalur cursor memakai `Repo.List`/`Repo.FindAfterCursor` yang mengabaikan
   `ListQuery`; `buildWhere` + `GetAllWithFilter` masih ada tapi tak terpakai.
   `model.ListQuery.Offset()` jadi legacy.
4. **`GET /api/v1/prestasi` masih offset pagination** + response bentuk lama
   (`{data,page,limit,total,total_pages,sort,order}`), belum cursor/negotiation.
5. **README bagian "Respons Error" & `helper/response.go` sudah usang** — format
   error sekarang `helper/errors.go` + `{success,code,message,fields}`.
   README juga belum menyebut endpoint `/users` dan `/api/v1/prestasi`.
6. **`CreateStudentRequest.IsActive` bertipe `bool` (bukan pointer)** → saat tidak
   dikirim, nilainya `false`, jadi default `TRUE` di kolom DB tidak terpakai.
   Pertimbangkan `*bool` + `omitnil` (bila modul baru menyinggung).
7. **`GET /students/:id` kehilangan `sort`/filter** — lihat #3.
8. Role `user` **tidak punya `student:create`**, jadi user biasa tidak dapat membuat
   mahasiswa "miliknya sendiri" lewat API; `owner_id` NULL hanya bisa diisi admin.
9. **Belum ada**: rate limiting, refresh token/logout, graceful shutdown,
   pagination pada prestasi, `POST /auth/register`, filter pada students.
   `helper.TooManyRequests` & `helper.ServiceUnavailable` sudah tersedia sebagai bekal.
10. `logs/app.log` ada, tapi logger hanya menulis ke stdout (`log.Printf`); tidak
    ada file writer yang terpasang.

---

## 13. Checklist Menambah Fitur Baru (urutan WAJIB)

1. **Migration** `migrations/00N_*.sql` — idempotent, ada komentar BLOCK-level
   (lihat gaya `004`).
2. **Model** di `app/model/` — entity + request struct + tag `validate:`.
3. **Repository** di `app/repository/` — parameterized, ORDER BY whitelist,
   error domain (`ErrNotFound`, `ErrDuplicate`, …).
4. **Rules** di `app/service/xxx_rules.go` — **tanpa import Fiber/DB**, + `xxx_rules_test.go`.
5. **Service** di `app/service/xxx_service.go` — handler: ParseBody → validasi →
   otorisasi → repo → `translateError`.
6. **Wiring** — `route/route.go` (route + `middleware.RequirePermission(...)`),
   `service.Services` bila service baru, `main.go` bila repo baru.
7. Validasi: `go build ./... && go vet ./... && go test ./...`
8. Update **README** (endpoint, env, skema) + **ai-usage.md** bila tugas menyangkut AI.
9. Kalau ada aturan bisnis murni baru: tambahkan tag custom di `helper/validator.go`
   **dan** `case` di `messageFor` (biar pesan Bahasa Indonesia).

### Gaya handler (pola baku service yang sudah ada)
```go
func (s XService) Handler(c *fiber.Ctx) error {
    var req model.XRequest
    if err := helper.ParseBody(c, &req); err != nil {
        return helper.BadRequest("body harus berupa JSON yang valid")
    }
    if errs := helper.ValidateStruct(req); len(errs) > 0 {
        return helper.Validation(errs)
    }
    // otorisasi bila perlu (CanAccessStudent / RequireOwnerOrAdmin)
    out, err := s.Repo.X(c.Context(), ...)
    if err != nil {
        return translateError(err, "mahasiswa")
    }
    return c.JSON(out) // atau c.Status(201).JSON(out)
}
```

---

## 14. Aturan untuk AI yang menerima handoff ini

1. Baca dokumen ini + `AGENTS.md` dulu; jangan menebak struktur.
2. **Tanya** bila tugas ambigu (bukan asumsi lalu revisi berkali-kali).
3. Jawab dalam **Bahasa Indonesia**; kode & identifier tetap bahasa Inggris.
4. Perubahan **minimal & terfokus**; jangan refactor yang tidak diminta.
5. Jangan ubah gaya kode yang ada (struct konkret, tanpa interface, `log`, bukan `slog`).
6. Selalu akhiri dengan `go build ./... && go vet ./... && go test ./...`.
7. Jelaskan singkat: file mana yang berubah, mengapa, dan cara mengetesnya.
8. Jangan pernah commit `.env`, `.exe`, `logs/`, `daftar tugas/`.
