# Panduan Memahami Proyek latihan-fiber

> Panduan ini ditulis untuk pembaca yang latar belakangnya PHP native / JavaScript
> sederhana, belum pernah pakai Go, dan belum paham clean architecture.
> Target akhir: paham alur aplikasi ini, bisa debug sendiri lewat Stack Overflow,
> dan bisa menambah endpoint baru dari nol sampai bisa dites di Postman — tanpa AI.

**Cara pakai panduan:** baca berurutan Bagian 1–5 sekali sambil membuka file yang
disebut di editor. Bagian 6–8 adalah referensi yang bisa dibuka kapan saja saat
kesulitan. Kalau kata/istilah asing muncul, cek dulu kamus di Bagian 8.

**Daftar isi**


| Bagian | Isi                                                                      | Kapan dibaca                |
| ------ | ------------------------------------------------------------------------ | --------------------------- |
| 1      | Dari PHP ke Go: peta konsep                                              | Sekali di awal              |
| 2      | Menyalakan aplikasi dari nol (Laragon → go run → health check)           | Sekali di awal              |
| 3      | Perjalanan satu request GET lewat Postman                                | Sekali di awal              |
| 4      | Anatomi setiap file + konsep Go yang dipakai                             | Pelan-pelan, file per file  |
| 5      | Membuat endpoint baru `/api/v1/students/:id/grades` dari nol             | Saat mau menambah fitur     |
| 6      | Memantau database tanpa phpMyAdmin (HeidiSQL/DBeaver + cheat sheet psql) | Kapan saja butuh lihat data |
| 7      | Debugging tanpa AI: baca error, baca log, cari di Stack Overflow         | Saat ada error              |
| 8      | Kamus istilah, padanan PHP→Go, sumber belajar lanjutan                   | Referensi                   |


Catatan penting: **seluruh kode contoh di Bagian 5 sudah diuji end-to-end**
(compile, migrasi database, dan 8 skenario request HTTP) sebelum ditulis ke
panduan ini. Kalau Anda mengikuti langkahnya persis, hasilnya akan sama.

---

## Bagian 1 — Dari PHP ke Go: Peta Konsep Baru

### 1.1 Hal paling membingungkan dulu: "Kok tidak ada Apache?"

Di dunia PHP, alur request-nya begini:

```
Browser → Apache (nginx) → interpreter PHP → MySQL → kembali ke Apache → Browser
```

Apache adalah *web server* yang berdiri di depan PHP. Ia yang mendengarkan
port 80/8080, lalu menyuruh PHP menjalankan file `.php` Anda.

Di dunia Go **tidak ada lapisan itu**. Program Go yang Anda tulis sendiri
berperan sebagai web server:

```
Browser → program Go Anda (Fiber) → PostgreSQL → kembali ke program Go → Browser
```

Fiber (framework yang dipakai proyek ini) membungkus *fasthttp*, web server
cepat yang tertanam langsung di dalam binary Go. Jadi saat menjalankan
`go run .`, program Anda sendiri yang "mendengarkan" port 3000.

Konsekuensinya untuk proyek ini:


| Komponen lama | Dipakai di proyek ini? | Keterangan                                      |
| ------------- | ---------------------- | ----------------------------------------------- |
| Apache        | **Tidak**              | Tidak ada file .php yang perlu di-serve         |
| nginx         | **Tidak**              | Program Go sudah jadi web server sendiri        |
| MySQL         | Tidak                  | Diganti PostgreSQL                              |
| phpMyAdmin    | **Tidak bisa dipakai** | phpMyAdmin hanya untuk MySQL/MariaDB            |
| PHP           | Tidak                  | Diganti Go                                      |
| PostgreSQL    | **Ya**                 | Satu-satunya service luar yang perlu dinyalakan |


Jadi di Laragon, yang perlu Anda nyalakan untuk proyek ini **hanya
PostgreSQL**. Tombol "Start All" pun tidak masalah (Apache ikut nyala tapi
tidak dipakai — biarkan saja, tidak mengganggu).

### 1.2 PostgreSQL itu service, bukan bagian dari Apache

MySQL dan PostgreSQL itu "pesaing" — keduanya *database server* mandiri yang
berjalan sebagai proses tersendiri, mendengarkan port sendiri:


| Software       | Peran                  | Port default         |
| -------------- | ---------------------- | -------------------- |
| Apache         | Web server (untuk PHP) | 80                   |
| nginx          | Web server (untuk PHP) | 80                   |
| **PostgreSQL** | **Database server**    | **5432**             |
| MySQL          | Database server        | 3306                 |
| Program Go ini | Web server + aplikasi  | 3000 (sesuai `.env`) |


MySQL biasanya Anda akses lewat phpMyAdmin (yang butuh Apache). PostgreSQL
tidak punya phpMyAdmin, tapi punya banyak alternatif GUI — lihat Bagian 6
(termasuk HeidiSQL yang **sudah terpasang di Laragon Anda**).

### 1.3 Padanan istilah PHP → Go


| Di PHP                             | Di Go (proyek ini)                  | Catatan                              |
| ---------------------------------- | ----------------------------------- | ------------------------------------ |
| `composer.json`                    | `go.mod`                            | daftar dependensi                    |
| `composer install`                 | `go mod download`                   | unduh dependensi                     |
| `vendor/`                          | cache modul Go                      | otomatis, tidak perlu diutak-atik    |
| namespace + `require`              | `package` + `import`                | lihat 4.9                            |
| class                              | `struct` + method                   | lihat 4.1                            |
| `new ClassName()`                  | `&Student{}` atau fungsi `New...()` | lihat 4.3                            |
| interface (eksplisit `implements`) | interface (implisit)                | lihat 4.3                            |
| PDO / mysqli                       | pgx                                 | driver database                      |
| `try {} catch {}`                  | `if err != nil {}`                  | Go mengembalikan error sebagai nilai |
| `null`                             | `nil`                               |                                      |
| `echo json_encode($x)`             | `c.JSON(x)` / helper.Success        |                                      |
| `$x->field`                        | `x.Field`                           |                                      |
| array asosiatif                    | `map[string]string`                 |                                      |
| `php file.php` jalankan langsung   | `go run .` compile dulu → jalankan  | binary .exe dihasilkan               |


### 1.4 Clean architecture versi proyek ini, dalam satu gambar

Ide utamanya sederhana: **satu file hanya pegang satu tanggung jawab**, dan
arah "siapa tahu siapa" diarahkan ke dalam.

```
             HTTP request (Postman)
                     │
                     ▼
┌──────────────────────────────────────────────┐
│  middleware/   requestid, recover, helmet,   │  penjaga di depan pintu
│                cors, RequestLogger            │
├──────────────────────────────────────────────┤
│  route/        daftar URL → siapa melayani   │  map saja, tanpa logika
├──────────────────────────────────────────────┤
│  app/service/  student_service.go            │  controller + use case:
│                (baca request, panggil rules, │  mengatur jalannya
│                 panggil repository, kirim    │  satu operasi bisnis
│                 response)                    │
│  app/service/  student_rules.go              │  business rules murni
│                (validasi, hitungan)          │  (TIDAK tahu HTTP/DB)
├──────────────────────────────────────────────┤
│  app/repository/ StudentRepository           │  satu-satunya yang
│                 (interface + implementasi    │  boleh menulis SQL
│                  PostgreSQL)                 │
├──────────────────────────────────────────────┤
│  database/     connection pool pgx           │
├──────────────────────────────────────────────┤
│  PostgreSQL (service luar, port 5432)        │
└──────────────────────────────────────────────┘
                     │
             helper/ = penerjemah masuk (baca request)
                      dan keluar (susun response)
             app/model/ = bentuk data (struct) yang dipakai semua layer
             config/    = perakitan + baca .env + logger
             main.go    = urutan perakitan saja
```

Aturan arah dependensinya (ini inti "clean"):

- `service` tahu `repository` lewat **interface**, bukan detail SQL-nya.
- `repository` TIDAK tahu `service`. `service` TIDAK tahu `route`.
- `student_rules.go` tidak mengimpor Fiber sama sekali → bisa dites tanpa
menyalakan server/database (itulah kenapa ada `student_rules_test.go`).

Analogi restoran: `route` adalah daftar menu, `service` adalah koki yang
mengatur resep, `student_rules.go` adalah buku resep, `repository` adalah
karyawan gudang yang satu-satunya boleh masuk gudang (database), `helper`
adalah pelayan yang menerjemahkan pesanan pelanggan dan menyajikan piring,
`model` adalah bentuk piring standar.

Kenapa repot dibagi begini? Dua keuntungan nyata di proyek ini:

1. **Bisa dites**: `go test ./...` menguji business rules tanpa database.
2. **Gampang dilacak saat error**: error 500 → lihat repository; error 422 →
 lihat rules; error routing/404 → lihat route. Tiap gejala punya satu
 ruang pencarian.

---

## Bagian 2 — Menyalakan Aplikasi dari Nol

### 2.1 Prasyarat (sekali saja)

1. Go terpasang — cek: `go version` (proyek ini memakai go 1.26.5).
2. Laragon terpasang, dengan PostgreSQL (ada di
 `C:\laragon\bin\postgresql\postgresql\bin`).
3. Dependensi proyek sudah diunduh: jalankan `go mod download` di folder
 proyek (perintah ini seperti `composer install`).

### 2.2 Langkah 1 — Nyalakan PostgreSQL

1. Buka Laragon → tombol **Start All** (atau klik kanan → PostgreSQL → Start).
 Yang benar-benar dibutuhkan hanya PostgreSQL, tapi Apache ikut nyala tidak
 masalah.
2. Verifikasi dari terminal (Cmder bawaan Laragon atau terminal lain):
  ```
   C:\laragon\bin\postgresql\postgresql\bin\pg_isready.exe -h localhost -p 5432
  ```

   Harus muncul: `localhost:5432 - accepting connections`.
   Kalau muncul "no response": PostgreSQL belum jalan — ulangi langkah 1.
  > Tips: tambahkan `C:\laragon\bin\postgresql\postgresql\bin` ke PATH
  > Windows agar bisa mengetik `psql` / `pg_isready` tanpa path panjang.
  > Cara cepat: di terminal Laragon biasanya path itu sudah tersedia.

### 2.3 Langkah 2 — Pastikan database &amp; tabelnya ada

Cek daftar database:

```
psql -U postgres -l
```

(masukkan password postgres bila diminta; di setup Laragon Anda password-nya
kosong). Cari baris `praktikum_backend` — nama database yang dipakai `.env`.

Kalau belum ada, buat + jalankan migrasi (satu kali saja):

```
psql -U postgres -c "CREATE DATABASE praktikum_backend;"
psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
```

Verifikasi tabelnya jadi:

```
psql -U postgres -d praktikum_backend -c "\dt"
```

Harus muncul tabel `students`. Kalau ingin cepat punya data contoh:

```
psql -U postgres -d praktikum_backend -f database/test_setup.sql
```

File itu mengisi 4 mahasiswa contoh. (Bagian 5 ada catatan penting: file ini
perlu diedit begitu Anda menambah tabel baru yang punya foreign key.)

### 2.4 Langkah 3 — Jalankan `go run .`

Di folder proyek (`C:\laragon\www\latihan-fiber`):

```
go run .
```

Apa yang sebenarnya terjadi di balik layar:

1. Go membaca `go.mod` untuk tahu versi Fiber/pgx yang dipakai.
2. Go **mengompilasi** seluruh kode Anda menjadi satu file `.exe` sementara
 (beda dengan PHP yang dibaca baris per baris saat request datang).
3. Windows menjalankan `.exe` itu; titik masuknya selalu fungsi `main()`
 di `main.go`.
4. Kalau kompilasi gagal, tidak ada yang jalan dan Anda langsung lihat
 error-nya di terminal (error kompilasi = salah ketik, import tidak
 terpakai, dsb — lihat Bagian 7).

### 2.5 Apa yang dikerjakan `main()` — langkah demi langkah

`main.go` sengaja hanya berisi urutan perakitan. Ini isinya dengan
penjelasan per blok:

```go
config.LoadEnv()
logger := config.NewLogger()
```

- `LoadEnv()` membaca file `.env` (memakai library godotenv) lalu menaruh
isinya ke *environment variable* — mirip variable `.env` di PHP, tapi
dibaca lewat `os.Getenv`.
- `NewLogger()` membuat logger JSON yang menulis ke DUA tempat sekaligus:
layar terminal DAN file `logs/app.log` (dirotasi otomatis saat mencapai
10 MB oleh library lumberjack). Semua aktivitas request tercatat di sini —
ini senjata utama debugging (Bagian 7).

```go
pool, err := database.NewPool(context.Background())
if err != nil {
    logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
    os.Exit(1)
}
defer pool.Close()
```

- `NewPool` merangkai alamat koneksi dari `.env`
(`postgres://postgres:@localhost:5432/praktikum_backend?sslmode=disable`),
membuat **connection pool** (kumpulan koneksi siap pakai — analoginya
kolam koneksi; query tinggal meminjam satu lalu mengembalikannya, lebih
cepat daripada membuka koneksi baru tiap request), lalu langsung `Ping`
untuk memastikan PostgreSQL benar-benar merespons (timeout 5 detik).
- Kalau gagal (PostgreSQL mati / password salah / database belum dibuat),
program **langsung berhenti** dengan kode exit 1. Ini by design:
aplikasi yang tanpa database tidak berguna, lebih baik gagal cepat.
- `defer pool.Close()` artinya "tutup pool saat fungsi main selesai" —
`defer` menjadwalkan perintah untuk dieksekusi belakangan, tepat sebelum
fungsi keluar.

```go
studentRepository := repository.NewStudentRepository(pool)
studentService := service.NewStudentService(studentRepository)
```

Perakitan dari dalam ke luar, seperti membangun rumah dari pondasi:

1. `NewStudentRepository(pool)` — buat pegawai gudang, beri akses pool.
2. `NewStudentService(studentRepository)` — buat koki, beri pegawai gudang.
 Perhatikan: yang diberikan adalah **interface** `StudentRepository`,
 bukan struct SQL-nya. Koki tahu "ada yang bisa FindByID/Create" tapi
 tidak tahu (dan tidak peduli) bahwa di belakangnya ada SQL PostgreSQL.

```go
app := config.NewApp(logger, pool, studentService)
port := config.GetEnv("APP_PORT", "3000")
```

`NewApp` (ada di `config/app.go`) membuat instance Fiber, memasang seluruh
middleware global, mendaftarkan semua route, lalu memasang penangkap 404
untuk URL tak dikenal. Setelah baris ini, "restoran" lengkap: pintu,
pelayan, menu, koki, gudang.

```go
go func() {
    if err := app.Listen(":" + port); err != nil { ... os.Exit(1) }
}()
```

`app.Listen(":3000")` = mulai mendengarkan port 3000 selamanya. Kata kunci
`go` di depannya membuatnya berjalan di **goroutine** (tugas ringan yang
berjalan paralel — seperti menyalakan mesin lalu tetap bisa mengerjakan hal
lain). Kalau ditulis tanpa `go`, baris ini akan menghentikan program di
situ selamanya dan baris setelahnya tidak pernah jalan.

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
```

Tiga baris ini = "tunggu sampai pengguna tekan Ctrl+C". `chan` (channel)
adalah pipa antar-goroutine; `<-quit` artinya blokir di sini menunggu
sesuatu masuk lewat pipa itu — dalam hal ini sinyal berhenti dari Windows.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := app.ShutdownWithContext(ctx); err != nil { ... }
```

Saat Ctrl+C ditekan, server tidak mati mendadak: ia diberi 10 detik untuk
menyelesaikan request yang sedang berjalan, baru benar-benar tutup. Ini
disebut *graceful shutdown*. Setelah itu barulah `defer pool.Close()` ikut
dijalankan (urutan defer kebalikan dari pemasangannya).

### 2.6 Langkah 4 — Verifikasi

Buka browser atau Postman ke:

```
GET http://localhost:3000/api/v1/health
```

Respons yang benar:

```json
{ "success": true, "message": "server dan database berjalan" }
```

Endpoint ini spesial: ia melakukan `pool.Ping` ke PostgreSQL, jadi kalau
pesan ini muncul berarti **keduanya** (program Go dan PostgreSQL) sehat.

### 2.7 Kalau gagal start — tabel gejala


| Pesan error di terminal                                | Artinya                                                              | Perbaikan                                       |
| ------------------------------------------------------ | -------------------------------------------------------------------- | ----------------------------------------------- |
| `failed to connect` / `connection refused` (port 5432) | PostgreSQL belum jalan                                               | Start lewat Laragon, cek `pg_isready`           |
| `password authentication failed for user "postgres"`   | `.env` DB_PASSWORD salah                                             | samakan dengan password PostgreSQL Anda         |
| `database "praktikum_backend" does not exist`          | langkah 2.3 dilewati                                                 | buat database + jalankan migrasi                |
| `listen tcp :3000: bind: Only one usage...`            | port 3000 dipakai program lain (kemungkinan server lama masih jalan) | tutup terminal server lama, atau ganti APP_PORT |
| `no required module provides package...`               | dependensi belum diunduh                                             | `go mod download`                               |
| error kompilasi `declared and not used: x`             | ada variable dibuat tapi tak dipakai (Go ketat soal ini)             | hapus variable tersebut                         |
| error kompilasi `imported and not used`                | ada import tak terpakai                                              | hapus baris importnya                           |


Lebih lengkap cara mendiagnosis: Bagian 7.

---

## Bagian 3 — Perjalanan Satu Request GET lewat Postman

### 3.1 Dulu bedah dulu: param, query, body, header — apa bedanya?

Postman punya beberapa tempat mengisi data, dan endpoint ini memakai semuanya
di tempat berbeda-beda. Ini konsep HTTP umum (berlaku juga di PHP):


| Tempat          | Bentuk                   | Contoh                           | Dipakai endpoint                          |
| --------------- | ------------------------ | -------------------------------- | ----------------------------------------- |
| **Path param**  | bagian dari URL          | `/students/1` → `id = 1`         | GET/PUT/PATCH/DELETE `:id`                |
| **Query param** | setelah tanda `?` di URL | `/students?page=2&limit=5`       | GET list                                  |
| **Body**        | isi kiriman (JSON)       | `{"nim":"...","name":"..."}`     | POST/PUT/PATCH                            |
| **Header**      | metadata kiriman         | `Content-Type: application/json` | POST/PUT/PATCH (wajib JSON di proyek ini) |


Aturan praktis proyek ini:

- **GET tidak pernah pakai body.** Semua inputnya lewat path/query param.
- **POST/PUT/PATCH wajib** set header `Content-Type: application/json`
(middleware `RequireJSON` menolak yang lain dengan 415) dan mengisi body
raw JSON.
- **DELETE tidak pakai body** dan suksesnya tidak mengembalikan body apa pun
(status 204 No Content).

### 3.2 Setup Postman untuk `GET /api/v1/students/1`

Di Postman:

1. Method: **GET** (dropdown di kiri kolom URL).
2. URL: `http://localhost:3000/api/v1/students/1`
3. Tab **Params**: kosong (angka `1` sudah menempel di URL sebagai path param).
4. Tab **Headers**: tidak wajib untuk GET.
5. Tab **Body**: **none** — GET tidak mengirim body.
6. Klik **Send**.

Respons 200 OK:

```json
{
  "success": true,
  "message": "mahasiswa ditemukan",
  "data": {
    "id": 1,
    "nim": "434241001",
    "name": "Ahmad Setiawan",
    "grade": 3.7,
    "is_active": true,
    "created_at": "2025-01-05T10:00:00+07:00"
  }
}
```

### 3.3 Perjalanan request tersebut di dalam kode — langkah demi langkah

Gambar besar dulu, lalu tiap nomor dijelaskan dengan potongan kode aslinya:

```
Postman
  │ ①  GET /api/v1/students/1
  ▼
Fiber web server (port 3000)
  │ ②  middleware: requestid → recover → helmet → cors → RequestLogger
  │ ③  middleware grup: RequireJSON (GET dilewati)
  │ ④  route: cocokkan "/students/:id" → studentService.Get
  ▼
service.Get (controller + use case)
  │ ⑤  helper.RequestContext  → context + timeout 5 detik
  │ ⑥  helper.ParamID         → baca "1" dari URL, ubah ke angka
  │ ⑦  s.repo.FindByID(ctx,1) → SQL ke PostgreSQL lewat pool
  ▼
repository.FindByID
  │ ⑧  QueryRow("SELECT ... WHERE id = $1").Scan(...)
  │ ⑨  hasil: struct Student  ATAU  ErrNotFound
  ▼
service kembali
  │ ⑩  translateError → pilih status HTTP (200 / 404 / 500)
  │ ⑪  helper.Success → susun JSON {success,message,data}
  ▼
middleware RequestLogger mencatat log  ← (② kembali arah)
Postman menerima JSON + status code
```

**① Request masuk.** Postman mengirim paket HTTP: method GET, path
`/api/v1/students/1`, header standar. Port 3000 sedang didengarkan oleh
program Go Anda.

**② Middleware global — penjaga berbaris.** Di `middleware/middleware.go`,
urutan pemasangan = urutan eksekusi:

```go
app.Use(requestid.New())       // beri setiap request satu ID unik
app.Use(recover.New())         // tangkap panic agar server tidak mati
app.Use(helmet.New())          // pasang header keamanan dasar
app.Use(cors.New())            // izinkan akses lintas domain (browser)
app.Use(RequestLogger(logger)) // catat setiap request ke log
```

Yang perlu Anda ingat: `recover` membuat server anti-mati (panic ditangkap
jadi respons 500, bukan server crash), dan `RequestLogger` mencatat semua
request ke `logs/app.log` lengkap dengan request_id, status, dan durasi.

**③ Middleware grup — RequireJSON.** Route students dipasang dengan
`middleware.RequireJSON`:

```go
students := api.Group("/students", middleware.RequireJSON)
```

Isinya hanya menolak POST/PUT/PATCH yang Content-Type-nya bukan JSON:

```go
if methodsWithBody[c.Method()] {   // GET tidak masuk kondisi ini
    ct := c.Get("Content-Type")
    if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
        return helper.Fail(c, fiber.StatusUnsupportedMediaType, ...)
    }
}
return c.Next()   // serahkan ke handler berikutnya
```

Untuk request GET kita, `methodsWithBody["GET"]` bernilai `false`, jadi
langsung `c.Next()`.

**④ Router mencocokkan.** Di `route/route.go`, daftar URL dipetakan ke
method service:

```go
students.Get("/:id", studentService.Get)
```

Fiber melihat `/api/v1/students/1` cocok dengan pola `/api/v1/students/:id`
— bagian `:id` menangkap nilai `1` dan disimpan sebagai parameter bernama
`id`. Fiber memanggil fungsi `studentService.Get` dengan `*fiber.Ctx`
(konteks request: berisi semua info HTTP masuk — header, param, body).

**⑤ Buat context + timeout.** Baris pertama hampir semua handler service:

```go
ctx, cancel := helper.RequestContext(c)
defer cancel()
```

`context.Context` adalah pembawa informasi lintas layer — di sini dipakai
untuk **batas waktu 5 detik** untuk operasi database. Kalau query lewat
dari 5 detik (misal database macet), context otomatis dibatalkan. `defer cancel()` memastikan pewaktu dibersihkan setelah handler selesai, apa pun
hasilnya. Perhatikan: yang dikirim ke repository adalah `ctx`, BUKAN
`*fiber.Ctx` — repository sengaja tidak tahu apa-apa soal HTTP.

**⑥ Baca path param.** `helper.ParamID(c)` membaca `c.Params("id")`
(nilainya `"1"`, masih string), mengubahnya ke angka dengan
`strconv.Atoi`, dan menolak angka di bawah 1. Gagal → 400 Bad Request.

**⑦ + ⑧ + ⑨ Repository menembak database.** Di
`app/repository/student_repository.go`:

```go
err := r.pool.QueryRow(
    ctx,
    `SELECT id, nim, name, grade, is_active, created_at
    FROM students
    WHERE id = $1`,
    id,
).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt)
```

- `$1` adalah placeholder parameter — **jangan pernah** menempel nilai
langsung ke string SQL (risiko SQL injection; ini padanan prepared
statement PDO).
- `QueryRow` = jalankan query, ambil satu baris.
- `.Scan(...)` = ambil kolom hasil query satu per satu ke field struct
(urutan kolom SELECT harus sama persis dengan urutan Scan).
- Kalau baris tidak ada, pgx mengembalikan error khusus yang diterjemahkan:

```go
if errors.Is(err, pgx.ErrNoRows) {
    return model.Student{}, ErrNotFound   // error sendiri: "data tidak ditemukan"
}
```

Kenapa tidak langsung kirim error pgx? Karena service tidak boleh tahu
detail PostgreSQL. Repository menerjemahkan error teknis menjadi error
domain (`ErrNotFound`, `ErrDuplicate`) — service tinggal memilih status
HTTP-nya.

**⑩ Service menerjemahkan error ke status HTTP.**

```go
student, err := s.repo.FindByID(ctx, id)
if err != nil {
    return translateError(c, err, "mahasiswa tidak ditemukan", "", "gagal mengambil mahasiswa")
}
```

`translateError` (ada di bagian bawah `student_service.go`) memetakan:
`ErrNotFound` → 404, `ErrDuplicate` → 409, sisanya → 500.

**⑪ Susun respons.** `helper.Success(c, 200, "mahasiswa ditemukan", student)`
menyusun struct `WebResponse` lalu `c.Status(200).JSON(...)` mengubahnya ke
JSON. Struct-nya di `app/model/student.go`:

```go
type WebResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
    Meta    *Meta  `json:"meta,omitempty"`
    Errors  any    `json:"errors,omitempty"`
}
```

Teks dalam tanda kutip balik itu **struct tag** `json:"..."` — menentukan
nama field di JSON hasil. `omitempty` = field tidak muncul kalau kosong,
makanya respons sukses tidak punya key `errors`.

**②(arah balik) RequestLogger menulis log.** Karena logger dipasang sebelum
handler, setelah handler selesai eksekusi kembali lewat dia — durasi
dihitung, lalu ditulis ke `logs/app.log` satu baris JSON. Buka file itu
setelah menerima request; di situlah semua jejaknya.

### 3.4 Cepat: endpoint lain, apa yang diisi di Postman

Semua base URL: `http://localhost:3000` — di collection Postman proyek ini
sudah ada variable `{{baseUrl}}` untuk itu.


| Endpoint      | Method | URL                  | Header                           | Body (raw JSON)                                                        |
| ------------- | ------ | -------------------- | -------------------------------- | ---------------------------------------------------------------------- |
| Daftar        | GET    | `/api/v1/students`   | —                                | —                                                                      |
| Detail        | GET    | `/api/v1/students/1` | —                                | —                                                                      |
| Tambah        | POST   | `/api/v1/students`   | `Content-Type: application/json` | `{"nim":"434241999","name":"Nama Baru","grade":3.25,"is_active":true}` |
| Ganti total   | PUT    | `/api/v1/students/1` | sama                             | semua field wajib: nim, name, grade, is_active                         |
| Ubah sebagian | PATCH  | `/api/v1/students/1` | sama                             | cukup yang diubah, contoh: `{"grade":3.9}`                             |
| Hapus         | DELETE | `/api/v1/students/1` | —                                | —                                                                      |


Query param GET `/api/v1/students` (isi di tab **Params**, Postman otomatis
menempelkannya ke URL):

```
?page=2&limit=5&search=ahmad&sort=name&order=desc&is_active=true
```

Semua opsional. `sort` hanya menerima: id, nim, name, grade, created_at.

**Kenapa body PATCH bisa sebagian?** Struct `PatchStudentRequest` memakai
pointer (`*string`, `*float64`):

```go
type PatchStudentRequest struct {
    NIM      *string  `json:"nim,omitempty"`
    Grade    *float64 `json:"grade,omitempty"`
    ...
}
```

Field yang tidak dikirim nilainya `nil` → kode tahu itu "tidak diubah".
Field yang dikirim nilainya alamat memori (`&value`) → kode ikuti alamatnya
dan ambil nilainya. (Konsep pointer dirangkum di 4.9.)

**Sukses vs gagal sekilas:**


| Status | Kapan                                                     | Bentuk body                                                                 |
| ------ | --------------------------------------------------------- | --------------------------------------------------------------------------- |
| 200    | GET/PUT/PATCH sukses                                      | `{success:true, message, data}`                                             |
| 201    | POST sukses                                               | sama + header `Location` URL data baru                                      |
| 204    | DELETE sukses                                             | **tidak ada body**                                                          |
| 400    | id bukan angka / body bukan JSON valid                    | `{success:false, message}`                                                  |
| 404    | id tidak ada di database / URL salah                      | `{success:false, message}`                                                  |
| 409    | NIM sudah dipakai (POST/PUT)                              | `{success:false, message}`                                                  |
| 415    | Content-Type bukan application/json (POST/PUT/PATCH)      | `{success:false, message}`                                                  |
| 422    | data lolos parsing tapi melanggar aturan (mis. grade 4.5) | `{success:false, message:"validasi gagal", errors:{field:pesan}}`           |
| 500    | error tak terduga (biasanya sisi database/kode)           | `{success:false, message:"terjadi error pada server"}` → cek `logs/app.log` |
| 503    | khusus /health saat database tidak bisa dihubungi         | `{success:false, message:"database tidak dapat dihubungi"}`                 |


---

## Bagian 4 — Anatomi Setiap File + Konsep Go yang Dipakainya

Baca pelan-pelan, sambil membuka file aslinya di editor. Urutan dibawah
mengikuti urutan data mengalir.

### 4.1 `app/model/student.go` — bentuk data (entity)

**Tanggung jawab:** mendefinisikan bentuk data yang dipakai semua layer.
Tidak ada logika kecuali hitungan kecil `Offset()`.

```go
type Student struct {
    ID        int       `json:"id"`
    NIM       string    `json:"nim"`
    Name      string    `json:"name"`
    Grade     float64   `json:"grade"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}
```

- `type X struct {...}` = kumpulan field dengan tipe tetap. Padanan paling
dekat di PHP: class dengan property bertipe, ATAU array asosiatif dengan
kunci tetap — tapi struct Go dikompilasi dan ditandai tipenya.
- `time.Time` = tipe tanggal-waktu bawaan Go.
- Tag `json:"is_active"`: penentu nama key saat struct diubah ke/dari JSON.

File ini juga berisi struct bentuk khusus: `CreateStudentRequest` (body
POST), `PatchStudentRequest` (body PATCH, pakai pointer agar bisa "tidak
dikirim"), `WebResponse` + `Meta` (bentuk respons), dan `ListQuery`
(hasil bacaan query param). Pisahnya begini membuat tiap kebutuhan punya
bentuk sendiri dan jelas.

### 4.2 `database/postgres.go` — pintu ke database

**Tanggung jawab:** membuat connection pool sekali di awal dan membagikannya.

```go
dsn := fmt.Sprintf(
    "postgres://%s:%s@%s:%s/%s?sslmode=%s",
    config.GetEnv("DB_USER", "postgres"),
    config.GetEnv("DB_PASSWORD", ""),
    ...
)
```

- DSN (alamat koneksi) dirangkai dari `.env`. `GetEnv(key, fallback)`
membaca environment variable; kalau kosong pakai nilai kedua.
- Konfigurasi pool: maksimal 10 koneksi, minimal 2, tiap koneksi dipakai
ulang maksimal 1 jam. Bayangkan kolam 10 koneksi; tiap query meminjam
satu lalu mengembalikan — jauh lebih murah daripada membuka koneksi baru
setiap request seperti `new PDO()` di setiap halaman PHP.
- Terakhir `pool.Ping` dengan timeout 5 detik — gagal ping = program
berhenti saat start dengan pesan jelas.

### 4.3 `app/repository/student_repository.go` — satu-satunya yang bicara SQL

**Tanggung jawab:** semua SQL hidup di sini. Tidak tahu apa pun soal HTTP.

Bagian pertama file: **kontrak (interface)**:

```go
type StudentRepository interface {
    FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
    FindByID(ctx context.Context, id int) (model.Student, error)
    Create(ctx context.Context, s model.Student) (model.Student, error)
    Update(ctx context.Context, s model.Student) (model.Student, error)
    Delete(ctx context.Context, id int) error
}
```

Interface = daftar kemampuan, tanpa implementasi. Bedanya dengan PHP:
struct di Go **tidak perlu menulis `implements`** — kalau method-nya cocok,
otomatis dianggap memenuhi interface. Inilah yang membuat service bisa
menerima "apa pun yang bisa FindByID/Create" tanpa peduli implementasinya.

Implementasi:

```go
type studentPostgresRepository struct {
    pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
    return &studentPostgresRepository{pool: pool}
}
```

- Struct kecil ini hanya menyimpan satu hal: pool. (Di PHP kira-kira:
`private pgxpool $pool`.)
- Fungsi `NewStudentRepository` adalah **constructor** versi Go — Go tidak
punya kata kunci constructor, jadi polanya selalu fungsi `New...` yang
mengembalikan instance.
- Nilai baliknya bertipe interface → pemanggil hanya melihat kontrak.

Method-nya (fungsi yang menempel pada struct — perhatikan `r *studentPostgresRepository` di dalam kurung, disebut *receiver*, padanan `$this` di PHP):

```go
func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
    ...
}
```

Pola penting yang dijumpai di file ini:

- `Query` → banyak baris (dikembalikan sebagai `rows`, dibaca satu-satu
dengan `for rows.Next()` + `Scan`), `QueryRow` → satu baris,
`Exec` → tanpa hasil (INSERT/UPDATE/DELETE).
- **Error domain:** `pgx.ErrNoRows` (baris tak ada) diterjemahkan jadi
`ErrNotFound`; kode PostgreSQL `23505` (unique violation, mis. NIM
kembar) diterjemahkan jadi `ErrDuplicate`. Sebaliknya, error tak dikenal
dibungkus dengan `fmt.Errorf("...: %w", err)` supaya jejak aslinya
tidak hilang (`%w` = *wrap*).
- `defer rows.Close()` = pastikan kursor ditutup apa pun yang terjadi.
- `FindAll` membangun WHERE secara dinamis (`buildFilter`) dengan
placeholder `$1, $2, ...` — nilai selalu terpisah dari teks SQL, aman
dari injection. LIMIT/OFFSET dihitung untuk pagination.

### 4.4 `app/service/student_service.go` — controller + use case

**Tanggung jawab:** mengatur satu operasi bisnis dari awal (baca request)
sampai akhir (kirim response). Di mata kuliah ini satu struct menanggung
dua peran: controller (pegang `*fiber.Ctx`) dan use case (panggil rules +
repository).

```go
type StudentService struct {
    repo repository.StudentRepository
}
```

Resep tiap handler hampir selalu lima langkah sama:

```go
func (s *StudentService) Get(c *fiber.Ctx) error {
    // 1. siapkan context berbatas waktu
    ctx, cancel := helper.RequestContext(c)
    defer cancel()

    // 2. baca & periksa input (path param / body / query)
    id, valid := helper.ParamID(c)
    if !valid {
        return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
    }

    // 3. jalankan business rules bila perlu (Create: ValidateCreate)
    // 4. panggil repository
    student, err := s.repo.FindByID(ctx, id)
    if err != nil {
        // 5. terjemahkan error → status HTTP
        return translateError(c, err, "mahasiswa tidak ditemukan", "", "gagal mengambil mahasiswa")
    }
    return helper.Success(c, fiber.StatusOK, "mahasiswa ditemukan", student)
}
```

`translateError` di akhir file memetakan error repository → status HTTP.
Pesan tidak-ditemukan/duplikat bisa berbeda per endpoint, makanya jadi
parameter. Alur PATCH agak lebih panjang: baca data lama (`FindByID`)
→ terapkan perubahan (`ApplyPatch`) → simpan (`Update`).

### 4.5 `app/service/student_rules.go` — business rules murni

**Tanggung jawab:** aturan bisnis (grade 0–4, NIM wajib isi, hitung total
halaman). File ini **tidak mengimpor Fiber dan database** — hanya struct
dan hitungan. Karena itu `student_rules_test.go` bisa mengujinya dengan
`go test` tanpa server dan tanpa database. Inilah contoh nyata manfaat
clean architecture: bagian paling berisiko salah (aturan) diuji termurah.

### 4.6 `helper/` — dua penerjemah kecil

- `request.go`: penerjemah **masuk**. `RequestContext` (timeout 5 dtk),
`ParamID` (baca+validasi `:id`), `ParseListQuery` (baca query param
dengan nilai bawaan aman + whitelist kolom sort — whitelist ini juga
penangkal injection di klausa ORDER BY).
- `response.go`: penerjemah **keluar**. `Success`, `SuccessList` (dengan
Meta pagination), `Created` (201 + header Location), `NoContent` (204),
`Fail`, `FailValidation` (422 + rincian error per field). Semua respons
lewat sini agar bentuknya konsisten.

### 4.7 `middleware/middleware.go` — penjaga di depan pintu

Lihat 3.3 untuk urutan eksekusinya. Dua hal menarik di file ini:

- **Closure** (fungsi yang mengembalikan fungsi) — pola memberi middleware
akses ke dependensi dari luar:

```go
func RequestLogger(logger *slog.Logger) fiber.Handler {   // terima logger
    return func(c *fiber.Ctx) error {                     // kembalikan handler
        ...
        err := c.Next()   // lanjut ke handler berikutnya (request maju)
        ...               // dieksekusi lagi saat response pulang
        return err
    }
}
```

- `c.Next()` = "silakan lanjut" — sebelum `c.Next()` kode berjalan saat
request masuk, setelahnya kode berjalan saat response keluar. Satu
middleware bisa bekerja di dua waktu.

### 4.8 `route/route.go`, `config/` — peta dan perakitan

- `route/route.go`: hanya daftar "URL ini dilayani fungsi itu". Grup
`/api/v1` → grup `/students` (+RequireJSON) → tiap method. Plus endpoint
`/health` yang langsung ping database.
- `config/app.go`: `NewApp` merakit semuanya — instance Fiber, middleware,
route, penangkap 404, dan `ErrorHandler` terakhir yang memformat error
tak tertangani + mencatatnya ke log.
- `config/env.go`: `LoadEnv` + `GetEnv`/`GetEnvInt` dengan nilai bawaan.
- `config/logger.go`: logger JSON ke layar + `logs/app.log` dengan rotasi
otomatis (lumberjack: 10 MB, simpan 5 file lama, kompres).

### 4.9 Cheat sheet pola Go yang muncul di mana-mana


| Pola              | Contoh                              | Artinya                                                                     |
| ----------------- | ----------------------------------- | --------------------------------------------------------------------------- |
| `:=`              | `x := 5`                            | buat variable + tebak tipenya (harus dipakai, kalau tidak: error kompilasi) |
| `if err != nil`   | ada di mana-mana                    | cara Go "try-catch": error adalah nilai balik yang diperiksa                |
| nilai balik ganda | `(student, err)`                    | satu fungsi boleh balik banyak nilai — pola `(hasil, error)` standar        |
| `defer x`         | `defer cancel()`                    | jalankan x saat fungsi keluar (pembersihan)                                 |
| `&x` dan `*T`     | `&req`, `*string`                   | `&` = ambil alamat; `*T` = tipe alamat/pointer; lihat PATCH di 3.4          |
| `nil`             | `req.NIM == nil`                    | "belum ada isinya" (versi Go dari null)                                     |
| receiver          | `func (s *StudentService) Get(...)` | method milik StudentService; `s` = padanan `$this`                          |
| huruf besar/kecil | `Student` vs `student`              | huruf besar = bisa diakses lintas package (publik); kecil = privat          |
| `any`             | `Data any`                          | tipe apa saja (padanan `mixed` di PHP)                                      |
| goroutine         | `go func(){...}()`                  | jalankan paralel ringan (dipakai di main.go untuk Listen)                   |
| package + import  | `package service`                   | satu folder = satu package; import lewat nama module `latihan-fiber/...`    |


---

## Bagian 5 — Membuat Endpoint Baru dari Nol: `/api/v1/students/:id/grades`

Studi kasus: **nilai mata kuliah milik satu mahasiswa** (sesuai keinginan
"endpoint nilai untuk mahasiswa dengan id tertentu"). Dua endpoint baru:


| Method | Path                        | Fungsi                                 |
| ------ | --------------------------- | -------------------------------------- |
| GET    | `/api/v1/students/1/grades` | daftar semua nilai mahasiswa id 1      |
| POST   | `/api/v1/students/1/grades` | tambah satu nilai untuk mahasiswa id 1 |


> Seluruh kode di bagian ini **sudah diuji nyata**: kompilasi bersih,
> migrasi jalan, dan 8 skenario request (sukses, 404, 415, 422) memberi
> hasil yang persis seperti ditulis di 5.7. Ikuti urutannya saja.

### 5.0 Rancang dulu di kertas, baru ketik kode

Sebelum menyentuh kode, jawab 5 pertanyaan ini (padanan "bikin ERD dulu"):

1. **Datanya apa?** → satu nilai: mata kuliah (`course`) dan angkanya
 (`score` 0–100), milik satu student.
2. **Tabelnya seperti apa?** → tabel `grades`: id, student_id (foreign key
 ke students — kalau mahasiswanya dihapus, nilainya ikut terhapus),
 course, score, created_at.
3. **Endpointnya apa saja?** → GET dan POST (di atas). DELETE/PUT bisa
 jadi latihan mandiri (5.9).
4. **Request POST bentuknya?** → `{"course":"Basis Data","score":92.5}`
 — student_id tidak diminta user, ia diambil dari URL.
5. **Error apa yang mungkin?** → student tidak ada (404), course kosong /
 score di luar 0–100 (422), bukan JSON (400/415).

Urutan pengerjaan selalu dari luar ke dalam — **database → model →
repository → service → route/wiring** — karena tiap lapisan memakai yang
di bawahnya.

### 5.1 Langkah 1 — Migrasi database (buat tabel `grades`)

Buat file baru `migrations/002_create_grades.sql`:

```sql
CREATE TABLE IF NOT EXISTS grades (
    id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course VARCHAR(100) NOT NULL,
    score DOUBLE PRECISION NOT NULL CHECK (score >= 0 AND score <= 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS grades_student_id_idx
ON grades (student_id);
```

Yang baru dibanding `001_create_students.sql`:

- `REFERENCES students(id) ON DELETE CASCADE` = **foreign key**: baris
nilai wajib menunjuk mahasiswa yang sungguh ada; kalau mahasiswa
dihapus, nilainya ikut terhapus (cascade).
- `CHECK (score >= 0 AND score <= 100)` = database ikut menjaga aturan,
lapis kedua setelah validasi di Go (defense in depth).
- Indeks pada `student_id`: karena pencarian selalu "semua nilai milik
student X".

Jalankan migrasinya:

```
psql -U postgres -d praktikum_backend -f migrations/002_create_grades.sql
```

**Cara tahu database &amp; tabelnya benar-benar jadi** (ini pertanyaan
penting Anda — jangan percaya begitu saja, verifikasi):

```
psql -U postgres -d praktikum_backend -c "\dt"        → harus ada students DAN grades
psql -U postgres -d praktikum_backend -c "\d grades"  → struktur kolom grades
```

`\d grades` akan menampilkan kolom, tipe, constraint, dan foreign key
("Foreign-key constraints: grades_student_id_fkey ... REFERENCES
students(id) ON DELETE CASCADE"). Kalau dua-duanya muncul, langkah 1 selesai.
(Cara yang sama juga bisa dilakukan lewat GUI — Bagian 6.)

**Satu catatan penting (temuan nyata saat panduan ini diuji):** file
`database/test_setup.sql` yang lama berisi `TRUNCATE TABLE students RESTART IDENTITY;` akan **error** begitu ada tabel lain yang menunjuk
students, karena PostgreSQL menolak mengosongkan tabel yang dirujuk.
Perbaiki baris itu menjadi:

```sql
TRUNCATE TABLE students, grades RESTART IDENTITY;
```

— sebut kedua tabel sekaligus. Ingat pelajaran umumnya: **setiap kali menambah
tabel ber-foreign-key, perbarui juga script seed/test**.

### 5.2 Langkah 2 — Model (bentuk datanya)

Buat file baru `app/model/grade.go`:

```go
package model

import "time"

type Grade struct {
	ID        int       `json:"id"`
	StudentID int       `json:"student_id"`
	Course    string    `json:"course"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateGradeRequest struct {
	Course string  `json:"course"`
	Score  float64 `json:"score"`
}
```

Pola sama seperti `student.go`: satu struct bentuk data tabel
(`Grade`), satu struct bentuk body POST (`CreateGradeRequest`). Tidak ada
`student_id` di request karena ia datang dari URL, bukan dari user.

### 5.3 Langkah 3 — Repository (SQL untuk grades)

Buat file baru `app/repository/grade_repository.go`:

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
)

type GradeRepository interface {
	FindByStudentID(ctx context.Context, studentID int) ([]model.Grade, error)
	Create(ctx context.Context, g model.Grade) (model.Grade, error)
}

type gradePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewGradeRepository(pool *pgxpool.Pool) GradeRepository {
	return &gradePostgresRepository{pool: pool}
}

func (r *gradePostgresRepository) FindByStudentID(
	ctx context.Context,
	studentID int,
) ([]model.Grade, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT id, student_id, course, score, created_at
		FROM grades
		WHERE student_id = $1
		ORDER BY id ASC`,
		studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar nilai: %w", err)
	}
	defer rows.Close()

	hasil := []model.Grade{}

	for rows.Next() {
		var g model.Grade

		if err := rows.Scan(
			&g.ID,
			&g.StudentID,
			&g.Course,
			&g.Score,
			&g.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("membaca baris nilai: %w", err)
		}

		hasil = append(hasil, g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, nil
}

func (r *gradePostgresRepository) Create(
	ctx context.Context,
	g model.Grade,
) (model.Grade, error) {
	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO grades (student_id, course, score)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		g.StudentID,
		g.Course,
		g.Score,
	).Scan(
		&g.ID,
		&g.CreatedAt,
	)

	if err != nil {
		// 23503 = foreign_key_violation: student_id tidak ada di tabel students
		if isForeignKeyViolation(err) {
			return model.Grade{}, ErrNotFound
		}

		return model.Grade{}, fmt.Errorf("menyimpan nilai: %w", err)
	}

	return g, nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}

	return false
}
```

Tiga hal yang **beda** dari repository students — pelajaran barunya:

1. **Filter by kolom, bukan id**: `WHERE student_id = $1` + `ORDER BY id  ASC` supaya urutannya stabil.
2. `**RETURNING id, created_at**`: INSERT di PostgreSQL bisa mengembalikan
 kolom yang baru dibuat database (id SERIAL dan created_at NOW() diisi
 otomatis oleh database) — kita tidak perlu menghitungnya sendiri.
3. **Kode error `23503`** (foreign key violation): kalau user POST nilai
 ke student id 999 yang tidak ada, INSERT gagal dengan kode ini. Kita
 terjemahkan menjadi `ErrNotFound` — error yang sama dengan yang dipakai
 repository students, sehingga service sudah tahu cara
 menerjemahkannya. (23505 = data duplikat; 23503 = rujukan tidak ada.)

### 5.4 Langkah 4 — Service (aturan + handler)

Dua file, meniru pola students: rules murni terpisah dari handler.

Buat file baru `app/service/grade_rules.go`:

```go
package service

import (
	"strings"

	"latihan-fiber/app/model"
)

// Aturan nilai mata kuliah: sama pola dengan student_rules.go,
// murni tanpa Fiber dan tanpa database sehingga mudah diuji.
func ValidateCreateGrade(req model.CreateGradeRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.Course) == "" {
		errs["course"] = "wajib diisi"
	}

	if req.Score < 0 || req.Score > 100 {
		errs["score"] = "harus bernilai antara 0 - 100"
	}

	return errs
}
```

Buat file baru `app/service/grade_service.go`:

```go
package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// GradeService menangani endpoint /api/v1/students/:id/grades.
// Ia memakai dua repository sekaligus: students untuk memastikan
// mahasiswanya ada, dan grades untuk datanya sendiri.
type GradeService struct {
	grades   repository.GradeRepository
	students repository.StudentRepository
}

func NewGradeService(
	grades repository.GradeRepository,
	students repository.StudentRepository,
) *GradeService {
	return &GradeService{grades: grades, students: students}
}

func (s *GradeService) ListByStudent(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// pastikan mahasiswanya ada, supaya balas 404 bukan daftar kosong
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translateError(c, err,
			"mahasiswa tidak ditemukan", "", "gagal mengambil mahasiswa")
	}

	items, err := s.grades.FindByStudentID(ctx, studentID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar nilai")
	}

	return helper.Success(c, fiber.StatusOK, "daftar nilai berhasil diambil", items)
}

func (s *GradeService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.CreateGradeRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Course = strings.TrimSpace(req.Course)

	if errs := ValidateCreateGrade(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	newGrade, err := s.grades.Create(ctx, model.Grade{
		StudentID: studentID,
		Course:    req.Course,
		Score:     req.Score,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}

		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menambahkan nilai")
	}

	return helper.Created(c, "nilai berhasil ditambahkan", newGrade,
		"/api/v1/students/"+strconv.Itoa(studentID)+"/grades")
}
```

Yang perlu diperhatikan di sini:

- **Dua repository dalam satu service** — hal yang normal: use case
"nilai mahasiswa" memang butuh cek students dan tulis grades.
- `translateError` TIDAK ditulis ulang — ia sudah ada di file
`student_service.go`, dan karena kedua file berada di **package yang
sama** (`service`), fungsinya langsung bisa dipakai. Inilah arti
"package" di Go: semua file di satu folder berbagi namespace.
- Handler Create memakai resep lima langkah yang sama persis dengan
4.4 (context → baca input → rules → repo → terjemahkan hasil).
- `c.BodyParser(&req)` mengisi struct dari body JSON — `&req` memberi
alamat struct agar isinya bisa diubah.

---

### 5.5 Langkah 5 — Menyambungkan semuanya (wiring)

Tiga file lama diedit sedikit. Ingat urutan sambungannya: repository dulu,
baru service, lalu daftarkan ke route lewat config.

**a) `main.go`** — tambah perakitan grade setelah student (kita butuh
`studentRepository` lagi, makanya variabelnya sudah dipisah):

```go
	// 3. Perakitan dari dalam ke luar: repository -> service
	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	gradeRepository := repository.NewGradeRepository(pool)
	gradeService := service.NewGradeService(gradeRepository, studentRepository)

	// 4. Aplikasi
	app := config.NewApp(logger, pool, studentService, gradeService)
```

(bagian lain `main.go` tidak berubah sama sekali)

**b) `config/app.go`** — `NewApp` menerima satu parameter tambahan:

```go
func NewApp(
	logger *slog.Logger, pool *pgxpool.Pool, studentService *service.StudentService,
	gradeService *service.GradeService,
) *fiber.App {
```

dan meneruskannya ke route:

```go
	middleware.Register(app, logger)
	route.Register(app, pool, studentService, gradeService)
```

**c) `route/route.go`** — signature `Register` ikut berubah, lalu
tambahkan grup route baru **di bawah** route students:

```go
func Register(
	app *fiber.App, pool *pgxpool.Pool,
	studentService *service.StudentService, gradeService *service.GradeService,
) {
```

```go
	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)

	// Endpoint baru: nilai milik satu mahasiswa.
	grades := students.Group("/:id/grades")
	grades.Get("/", gradeService.ListByStudent)
	grades.Post("/", gradeService.Create)
```

`students.Group("/:id/grades")` = sub-route dari `/students`, jadi URL
lengkapnya `/api/v1/students/:id/grades`, dan middleware RequireJSON grup
students otomatis berlaku juga (POST tanpa header JSON akan ditolak 415).

Kenapa tiga file ini harus ikut diedit? Karena desain proyek ini sengaja
**tidak ada magic**: semua dependensi disuntik lewat parameter (istilahnya
*dependency injection manual*). Tidak ada auto-register seperti framework
PHP yang memindai folder controller. Keuntungannya: semua hubungan
kelihatan di satu tempat (`main.go`), tidak ada yang tersembunyi.

### 5.6 Cek kompilasi sebelum menyalakan server

```
go build ./...
```

- Tidak keluar apa-apa = sukses (di Go, diam = berhasil).
- `go vet ./...` juga berguna: pemeriksa statis tambahan.

Kalau muncul `undefined: service.GradeService` → service belum tersambung
di route/config; `imported and not used` → ada import kelebihan. Setiap
error kompilasi selalu menyebut file dan nomor baris — baca dari atas.

Lalu jalankan seperti biasa: `go run .` dan cek `GET /api/v1/health`.

### 5.7 Langkah 6 — Tes di Postman

Buat request baru. Isi persis seperti tabel ini:


| #   | Method | URL                                              | Header                           | Body                                         |
| --- | ------ | ------------------------------------------------ | -------------------------------- | -------------------------------------------- |
| 1   | GET    | `http://localhost:3000/api/v1/students/1/grades` | —                                | none                                         |
| 2   | POST   | `http://localhost:3000/api/v1/students/1/grades` | `Content-Type: application/json` | `{"course":"Basis Data","score":92.5}` (raw) |
| 3   | POST   | sama                                             | sama                             | `{"course":"Jaringan Komputer","score":88}`  |
| 4   | GET    | sama                                             | —                                | none                                         |
| 5   | GET    | `.../students/999/grades`                        | —                                | none                                         |
| 6   | POST   | `.../students/1/grades`                          | sama                             | `{"course":"","score":150}`                  |
| 7   | POST   | `.../students/1/grades`                          | TANPA Content-Type               | form-data apa saja                           |
| 8   | POST   | `.../students/999/grades`                        | `Content-Type: application/json` | `{"course":"X","score":50}`                  |


Hasil yang benar (hasil uji nyata saat panduan ini dibuat):

1. **200** — `{"success":true,"message":"daftar nilai berhasil diambil","data":[]}`
 (masih kosong karena baru pertama kali; kalau student 1 tidak ada,
 pastikan `test_setup.sql` sudah dijalankan)
2. **201** —
 `{"success":true,"message":"nilai berhasil ditambahkan","data":{"id":1,"student_id":1,"course":"Basis Data","score":92.5,"created_at":"..."}}`
 plus header `Location: /api/v1/students/1/grades`
3. **201** — id 2, Jaringan Komputer, score 88
4. **200** — `data` sekarang berisi dua objek nilai tadi
5. **404** — `{"success":false,"message":"mahasiswa tidak ditemukan"}` ← bukti
 cek keberadaan student di ListByStudent bekerja
6. **422** — `{"success":false,"message":"validasi gagal","errors":{"course":"wajib diisi","score":"harus bernilai antara 0 - 100"}}`
7. **415** — Content-Type harus application/json (middleware RequireJSON)
8. **404** — INSERT ditolak foreign key (23503) → diterjemahkan jadi
 "mahasiswa tidak ditemukan"

Kalau semua lolos: **selamat, endpoint baru Anda hidup.** Verifikasi
terakhir di sisi data — cek isinya langsung di database:

```
psql -U postgres -d praktikum_backend -c "SELECT * FROM grades;"
```

atau lewat GUI (Bagian 6): refresh tabel `grades`, dua baris tadi harus
terlihat. Ini membulatkan pemahaman: Postman → program Go → PostgreSQL →
kembali.

### 5.8 Checklist: "mau ubah/expand apa → file mana yang disentuh"


| Keinginan                               | File yang disentuh                                                        |
| --------------------------------------- | ------------------------------------------------------------------------- |
| Tabel/kolom database baru               | `migrations/00X_....sql` (+ jalankan psql)                                |
| Bentuk data/request/response baru       | `app/model/....go`                                                        |
| Query SQL baru/diubah                   | `app/repository/....go` (interface + implementasi)                        |
| Aturan bisnis baru (validasi, hitungan) | `app/service/xxx_rules.go`                                                |
| Endpoint/handler baru                   | `app/service/xxx_service.go`                                              |
| Mendaftarkan URL baru                   | `route/route.go` (+ teruskan service lewat `config/app.go` dan `main.go`) |
| Ubah port/koneksi/log                   | `.env`                                                                    |
| Middleware global baru                  | `middleware/middleware.go`                                                |


Urutan aman mengerjakan selalu: **database → model → repository → rules →
service → wiring → build → tes Postman → cek data**.

### 5.9 Latihan mandiri (tanpa AI!)

Kerjakan berurutan, masing-masing hanya butuh pola yang sudah Anda lihat:

1. **DELETE `/api/v1/students/:id/grades/:gradeId`** — butuh `c.Params`
 dua parameter dan method `Delete` baru di repository. Uji: 204 saat
 ada, 404 saat tidak.
2. **GET rata-rata**: `/api/v1/students/:id/grades/average` — SQL
 `AVG(score)` di repository; tampilkan hasil di `data`. (Tantangan
 kecil: `AVG` mengembalikan `NULL` kalau kosong — coba tangani.)
3. **GET `/api/v1/grades?limit=...`** — nilai semua mahasiswa dengan
 JOIN (`JOIN students ON students.id = grades.student_id`) dan
 pagination. Setelah ini Anda praktis sudah menulis ulang seluruh
 pola proyek dengan tangan sendiri.

Setiap selesai satu: `go build ./...` dulu, baru tes Postman, baru lihat
data di GUI. Kalau macet: Bagian 7.

---

## Bagian 6 — Memantau Database Tanpa phpMyAdmin

### 6.1 Kenapa phpMyAdmin tidak bisa dipakai di sini

phpMyAdmin adalah aplikasi **PHP yang berjalan di Apache** dan hanya bisa
berbicara dengan **MySQL/MariaDB**. Proyek ini tidak punya Apache dan tidak
memakai MySQL — jadi phpMyAdmin memang tidak relevan. Supabase yang dulu
Anda pakai adalah PostgreSQL juga (versi cloud dengan dashboard), jadi
yang Anda butuhkan sekarang hanya "Supabase-nya versi lokal": **GUI untuk
PostgreSQL**. Pilihannya banyak dan gratis.

### 6.2 Dulu cara manual: memastikan PostgreSQL hidup dan isinya benar

Tanpa GUI apa pun pun Anda bisa memantau semuanya. Ini penting dipahami
dulu supaya GUI hanya jadi kemudahan, bukan sihir.

**PostgreSQL-nya jalan tidak?**

```
C:\laragon\bin\postgresql\postgresql\bin\pg_isready.exe -h localhost -p 5432
```

```
tasklist | findstr postgres     → daftar proses postgres (ada beberapa: normal)
```

**Ada database apa saja?** (setara melihat daftar database di phpMyAdmin)

```
psql -U postgres -l
```

**Masuk ke database proyek** (semua perintah berikut dijalankan setelah ini):

```
psql -U postgres -d praktikum_backend
```

Prompt berubah menjadi `praktikum_backend=#` — sekarang Anda "berada di
dalam" database, seperti klik nama database di phpMyAdmin.

### 6.3 Cheat sheet psql (pengganti klik-klik phpMyAdmin)


| Perintah                                                                | Setara phpMyAdmin    | Keterangan                        |
| ----------------------------------------------------------------------- | -------------------- | --------------------------------- |
| `\dt`                                                                   | daftar tabel di kiri | daftar tabel database aktif       |
| `\d students`                                                           | tab Structure        | struktur kolom + constraint tabel |
| `\di`                                                                   | daftar index         |                                   |
| `\x` (lalu SELECT)                                                      | tampilan vertikal    | nyaman untuk 1 baris panjang      |
| `SELECT * FROM students LIMIT 5;`                                       | tab Browse           | lihat isi tabel                   |
| `SELECT id,nim,name FROM students WHERE is_active = true ORDER BY id;`  | pencarian + filter   |                                   |
| `INSERT INTO students (nim,name,grade) VALUES ('434200001','Tes',3.0);` | tab Insert           |                                   |
| `UPDATE students SET grade = 3.5 WHERE id = 2;`                         | edit baris           | selalu pakai WHERE!               |
| `DELETE FROM students WHERE id = 99;`                                   | hapus baris          | selalu pakai WHERE!!              |
| `SELECT COUNT(*) FROM students;`                                        | jumlah baris         |                                   |
| `\q`                                                                    | keluar               | quit                              |


Aturan psql: perintah mulai `\` tanpa titik-koma; SQL biasa **wajib
diakhiri titik-koma**. Kalau sepertinya "hang", biasanya titik-komanya
kurang — ketik `;` lalu Enter.

### 6.4 GUI gratis untuk PostgreSQL


| GUI                                             | Gratis?               | Ada di komputer Anda?                 | Karakter                                                  |
| ----------------------------------------------- | --------------------- | ------------------------------------- | --------------------------------------------------------- |
| **HeidiSQL**                                    | ya                    | **SUDAH** — `C:\laragon\bin\heidisql` | ringan, cepat dibuka, fitur pas untuk belajar             |
| **pgAdmin 4**                                   | ya (resmi PostgreSQL) | belum                                 | paling lengkap, agak berat (jalankan web server sendiri)  |
| **DBeaver Community**                           | ya                    | belum                                 | mendukung banyak database sekaligus, tampilan seperti IDE |
| VS Code + ekstensi Database Client / PostgreSQL | ya (sebagian)         | belum                                 | kalau sudah betah di VS Code                              |


Rekomendasi untuk Anda: **mulai dari HeidiSQL** karena sudah terpasang di
Laragon dan tidak perlu unduh apa pun. Kalau nanti butuh yang lebih kuat,
pasang DBeaver Community (dbeaver.io) atau pgAdmin (postgresql.org/ftp/pgadmin).

### 6.5 Langkah konek HeidiSQL ke PostgreSQL proyek ini

1. Buka HeidiSQL (Laragon → kanan → HeidiSQL, atau jalankan
 `heidisql.exe` dari folder itu).
2. Kiri bawah tombol **New** → pilih **Network type: PostgreSQL (TCP/IP)**
 (ini kuncinya — HeidiSQL bisa MySQL DAN PostgreSQL).
3. Isi:
  
  | Field         | Nilai                                                 |
  | ------------- | ----------------------------------------------------- |
  | Hostname / IP | `localhost` (atau `127.0.0.1`)                        |
  | Port          | `5432`                                                |
  | User          | `postgres`                                            |
  | Password      | password postgres Anda (kosong di setup Laragon Anda) |
  | Database      | `praktikum_backend`                                   |
  
4. **Open**. Kalau konek, panel kiri menampilkan daftar tabel:
 `students` (dan `grades` setelah Bagian 5).
5. Klik tabel → tab **Data** = Browse phpMyAdmin; klik kanan tabel →
 **Create new** → Table = bikin tabel lewat GUI; tab **Query** = tempat
 menjalankan SQL apa pun (bisa dipakai untuk menjalankan isi file
 migrasi juga — copy-paste SQL-nya ke situ lalu F9).

DBeaver versi singkat: Database → New Connection → PostgreSQL →
localhost:5432, user postgres, database praktikum_backend → Finish. Di
panel kiri: Databases → praktikum_backend → public → Tables.

### 6.6 Kebiasaan pemantauan yang sehat saat develop

1. Buka GUI (atau psql) **di sebelah Postman** saat mengembangkan.
2. Setelah POST sukses → refresh tabel di GUI → data baru harus muncul.
 Ini membuktikan end-to-end jalan, bukan cuma status 201.
3. Setelah DELETE → refresh lagi → baris hilang.
4. Kalau respons 500 → cek `logs/app.log` (pesan asli errornya ada di
 sana), baru cek data di GUI.
5. Database "aneh"? Bandingkan struktur di GUI dengan `\d nama_tabel` —
 kadang migrasi belum dijalankan, atau dijalankan ke database yang salah
 (ingat: PostgreSQL bisa punya banyak database; `.env` DB_NAME menentukan
 yang dipakai program).

---

## Bagian 7 — Debugging Tanpa AI (dengan Stack Overflow)

### 7.1 Urutan kerja saat ada error — jangan langsung cari-cari

1. **Baca pesan errornya sampai selesai.** Error Go selalu berformat
 `file.go:42: pesan`. Artinya: file `file.go`, baris 42. Klik/buka baris
 itu — 80% kasus penyebabnya sudah kelihatan.
2. **Tentukan lapisannya dulu**, biar pencarian sempit:
  - Gagal saat `go run` / `go build` → error **kompilasi** (salah ketik,
   import, tipe).
  - Respons 404/400/415/422 → bukan bug, itu **aturan yang bekerja**;
  bandingkan request Anda dengan Bagian 3.
  - Respons 500 → error di **jalur database/kode** → buka `logs/app.log`.
  - Server tidak bisa distart / health 503 → **masalah environment**
  (PostgreSQL, `.env`, port).
3. **Buka `logs/app.log`** — semua request dan error tak tertangani
 tercatat di sana dengan path, status, dan pesan aslinya. Karena
 formatnya JSON satu baris per kejadian, cari `"level":"ERROR"` atau
 `"status":500`.
4. **Pisahkan variabel:** ubah satu hal → jalankan ulang → lihat apakah
 gejalanya berubah. (Seperti menutup kolom air satu-satu, bukan
 membongkar semua sekaligus.)

### 7.2 Tabel error umum → penyebab → perbaikan


| Pesan/gejala                                     | Penyebab                                            | Perbaikan                                          |
| ------------------------------------------------ | --------------------------------------------------- | -------------------------------------------------- |
| `connection refused` (port 5432)                 | PostgreSQL mati                                     | start via Laragon, cek pg_isready                  |
| `password authentication failed`                 | DB_PASSWORD di .env salah                           | perbaiki .env, restart server                      |
| `database "x" does not exist`                    | DB_NAME salah / belum dibuat                        | buat database atau perbaiki .env                   |
| `relation "students" does not exist`             | tabel belum dibuat (migrasi belum jalan)            | jalankan file migrasi (2.3)                        |
| `column "x" does not exist`                      | kolom di SQL ≠ kolom tabel                          | bandingkan `\d tabel` dengan SELECT-nya            |
| `syntax error at or near...`                     | SQL salah tulis                                     | cek koma/kutip di query, jalankan dulu di GUI/psql |
| `duplicate key value violates unique constraint` | data duplikat (NIM kembar)                          | ini 409 by design; cek data                        |
| `Only one usage of each socket address`          | port 3000 dipakai lain                              | matikan server lama / ganti APP_PORT               |
| `no required module provides package`            | dependensi belum diunduh                            | `go mod download`                                  |
| `declared and not used: x`                       | variabel tak terpakai                               | hapus variabel x                                   |
| `imported and not used: "..."`                   | import tak terpakai                                 | hapus import itu                                   |
| `undefined: X`                                   | nama salah / file baru belum disimpan / lupa wiring | cek ejaan, simpan file, cek Bagian 5.5             |
| `cannot use x (type A) as type B`                | tipe tidak cocok                                    | biasanya lupa `&`, atau salah tipe di Scan         |
| `invalid memory address or nil pointer`          | pakai pointer nil                                   | biasanya field PATCH bernilai nil tapi diakses `*` |
| panic + server tetap hidup                       | panic di handler, ditangkap recover                 | baca stack trace (7.5)                             |
| Postman "Could not get any response"             | server mati / port salah                            | cek server jalan &amp; URL                         |


### 7.3 Cara memakai Stack Overflow dengan benar

Kunci mencari di SO: **ganti emosi dengan kata kunci**. Resepnya:

1. **Ambil frasa error yang paling spesifik**, buang bagian yang spesifik
 milik Anda. Contoh: error lengkapnya
 `pq: relation "students" does not exist` → cari:
 `postgresql relation does not exist`. (bagian nama tabel disingkirkan.)
2. **Tambahkan nama teknologinya**: `go pgx connection refused`,
 `fiber bodyparser cannot parse`, bukan cuma "golang error".
3. **Baca jawaban yang dicentang hijau (accepted) dulu**, lalu urutkan
 berdasarkan suara. Tapi perhatikan **tanggalnya**: jawaban 2016 tentang
 Fiber v1 sering sudah salah untuk v2 — bandingkan dengan versi di
 `go.mod`.
4. **Skim 3–5 hasil teratas** sebelum mencoba apa pun. Banyak jawaban
 menawarkan beberapa solusi; pilih yang paling mirip konteks Anda.
5. **Cocokkan gejala persis**: kalau orang itu errornya saat `go build`
 tapi Anda errornya saat request masuk — itu beda masalah, lanjut cari.
6. Setelah menemukan kandidat solusi: **pahami satu baris yang diubah,
 jangan tempel buta.** Ubah minimal, jalankan, amati.

Contoh nyata alurnya — kasus `connection refused` saat `go run .`:

1. Pesan: `failed to connect to "host=localhost port=5432": ... connection refused`.
2. Analisis sendiri dulu (7.1): errornya terjadi SEBELUM server jalan,
 saat `NewPool` ping → koneksi ke PostgreSQL gagal → masalah environment.
3. Cek cepat: `pg_isready -h localhost -p 5432` → ternyata "no response".
4. Baru kalau bingung kenapa Laragon bilang PostgreSQL jalan padahal
 tidak: cari SO `laragon postgresql not starting` atau
 `pg_isready no response windows`.
5. Hasil SO khas: PostgreSQL di Laragon perlu di-start dari menu Laragon,
 bukan sekadar folder ada; atau port diganti oleh instalasi PostgreSQL
 lain (dua PostgreSQL terpasang bentrok port 5432). Cek keduanya, satu
 per satu.

### 7.4 Sebelum bertanya (kalau benar-benar tidak ketemu)

SO menolak pertanyaan yang buruk. Sebelum bertanya, siapkan **MRE**
(minimal reproducible example): potongan kode terkecil yang kalau
dijalankan orang lain, errornya muncul; plus teks error lengkap; plus
apa yang sudah Anda coba. Dalam proses memangkas kode jadi MRE itu,
sangat sering penyebabnya ketemu sendiri. Ini bukan kebetulan — itulah
cara kerja pemangkasan masalah.

### 7.5 Alat debugging bawaan yang cukup untuk proyek ini

- **Cetak variabel**: `fmt.Println("DEBUG nilai id:", id)` sementara,
hapus setelah ketemu. Untuk struct: `fmt.Printf("%+v\n", student)`.
- **Log sementara di logger proyek**: `logger.Info("debug", slog.Int("id", id))`
→ muncul di layar + app.log dengan request_id.
- **Baca stack trace panic**: saat panic, recover middleware menangkapnya
tapi log tetap memuat jejaknya — baris PERTAMA bertanda `panic:` adalah
pesan, baris pertama `latihan-fiber/...` di bawahnya adalah titik
kejadian di kode Anda (abaikan baris framework/fasthttp di antaranya).
- **Baca dokumentasi resmi dari terminal**: `go doc fiber.Ctx.Params`,
`go doc pgxpool.Pool.QueryRow` — cepat, tanpa internet. Versi webnya:
pkg.go.dev (dokumentasi paket Go), docs.gofiber.io (Fiber),
pkg.go.dev/github.com/jackc/pgx/v5 (pgx), postgresql.org/docs (SQL).

### 7.6 Kapan percaya siapa


| Sumber                                          | Kapan dipercaya                   | Hati-hati                               |
| ----------------------------------------------- | --------------------------------- | --------------------------------------- |
| Pesan error compiler Go                         | selalu — Go compiler jarang salah | pesan kadang panjang; baca yang pertama |
| Dokumentasi resmi (pkg.go.dev, docs.gofiber.io) | selalu                            | pastikan versi cocok go.mod             |
| Jawaban SO accepted + suara tinggi              | umumnya                           | cek tahun vs versi library Anda         |
| Jawaban SO suara rendah/tanpa centang           | coba dulu di kode mainan          | jangan langsung ke produksi             |
| Blog/tutorial                                   | sebagai wawasan                   | sering tidak mutakhir                   |


---

## Bagian 8 — Kamus Istilah &amp; Jalan Belajar Selanjutnya

### 8.1 Kamus istilah proyek ini


| Istilah                | Arti singkat                                                        |
| ---------------------- | ------------------------------------------------------------------- |
| API / REST API         | program yang dikontrol lewat HTTP (GET/POST/PUT/PATCH/DELETE)       |
| endpoint               | satu alamat + method, mis. `GET /api/v1/students/1`                 |
| handler                | fungsi yang melayani satu endpoint                                  |
| JSON                   | format teks data yang dipakai body request/response                 |
| struct                 | tipe bentuk data buatan (field bernama bertipe) — "class"-nya Go    |
| interface              | daftar kemampuan (kontrak) tanpa implementasi                       |
| repository             | lapisan yang berisi semua SQL                                       |
| service / use case     | lapisan aturan bisnis + pengatur alur                               |
| middleware             | kode yang mengawal request sebelum/sesudah handler                  |
| migration              | file SQL perubahan struktur database, dijalankan berurutan          |
| seed                   | data contoh untuk development/test (test_setup.sql)                 |
| pool (connection pool) | kumpulan koneksi database siap pinjam                               |
| DSN                    | alamat koneksi database (`postgres://user:pass@host:port/db`)       |
| goroutine              | tugas paralel ringan (`go ...`)                                     |
| context.Context        | pembawa batas waktu/pembatalan lintas lapisan                       |
| panic                  | error fatal yang menghentikan eksekusi (ditangkap recover)          |
| dependency injection   | dependensi dikirim lewat parameter/fungsi New, bukan dibuat sendiri |
| graceful shutdown      | mati server dengan menyelesaikan request berjalan dulu              |
| FK / foreign key       | kolom yang wajib menunjuk baris tabel lain                          |
| ILIKE                  | LIKE versi PostgreSQL yang tidak peduli huruf besar/kecil           |
| pagination             | membagi daftar panjang per halaman (page/limit/Meta)                |


### 8.2 Padanan cepat PHP → Go (lanjutan 1.3)


| Kebutuhan           | PHP                                             | Go (pola proyek ini)                  |
| ------------------- | ----------------------------------------------- | ------------------------------------- |
| route               | `index.php?a=students/edit` / framework         | `route/route.go` eksplisit            |
| ambil body JSON     | `json_decode(file_get_contents('php://input'))` | `c.BodyParser(&req)`                  |
| kirim JSON          | `header(...); echo json_encode($x)`             | `helper.Success(...)`                 |
| query parameterized | PDO `?` + execute                               | pgx `$1, $2` + args                   |
| ambil 1 baris       | `fetch()` + `if(!$row)`                         | `QueryRow().Scan()` + `pgx.ErrNoRows` |
| banyak baris        | `while($row = fetch())`                         | `for rows.Next() { rows.Scan() }`     |
| affected rows       | `$stmt->rowCount()`                             | `tag.RowsAffected()`                  |
| environment         | `$_ENV` / `getenv()`                            | `config.GetEnv` + godotenv            |
| unit test           | phpunit (opsional)                              | `go test ./...` bawaan                |


### 8.3 Status HTTP yang dipakai proyek ini

2xx sukses (200 OK, 201 Created, 204 No Content) · 4xx salah klien
(400 format salah, 404 tidak ada, 409 konflik/duplikat, 415 tipe media
salah, 422 isi melanggar aturan) · 5xx salah server (500 internal,
503 layanan tak siap).

### 8.4 Jalur belajar selanjutnya yang pas untuk proyek ini

1. **go.dev/tour** — dasar sintaks Go interaktif (2–3 jam, kerjakan
 sampai pointer/method/interface).
2. Baca ulang Bagian 4 sambil membuka tiap file — coba jelaskan tiap
 baris dengan suara sendiri.
3. Kerjakan 5.9 dengan tangan sendiri, tanpa menyalin.
4. **Effective Go** (go.dev/doc/effective_go) — cara idiomatik menulis
 Go, baca setelah nyaman dengan sintaks.
5. docs.gofiber.io untuk hal baru di Fiber, pkg.go.dev untuk paket apa pun.
6. Baru setelah itu (opsional): `go test` lebih dalam, transaksi SQL
 (`pool.Begin`), autentikasi JWT, dsb.

### 8.5 Ringkasan satu paragraf

Program ini adalah satu binary Go yang **menjadi web server sendiri**
(Fiber) dan berbicara ke **PostgreSQL** lewat pool pgx. Request masuk
melalui middleware, dicocokkan di route, dilayani service yang memakai
business rules murni + repository pembicara SQL, lalu respons disusun
helper agar bentuknya konsisten. Semua bagian dirakit manual di main.go —
tidak ada yang tersembunyi. Untuk menambah endpoint: buat tabel (migrasi)
→ model → repository → rules → service → sambungkan di route/config/main
→ build → tes di Postman → verifikasi datanya lewat psql atau HeidiSQL.
Saat macet: baca error (file:baris), baca `logs/app.log`, pisahkan
lapisannya, cari di Stack Overflow dengan frasa error + nama library.

*Panduan dibuat sebagai bagian dari Tugas 4 — Clean Architecture.*