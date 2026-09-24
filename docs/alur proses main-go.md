 Sudah saya baca source-nya. Berikut jawaban to the point, urut sesuai
 pertanyaanmu.

 ─────────────────────────────────────────────────────────────────────

 1. config.LoadEnv() — config/env.go:24

 Tipe data:

 ```go
data, err := os.ReadFile(".env")
 ```

 - := = short variable declaration (Go menebak tipe sendiri).
 - os.ReadFile (Go 1.16+, pengganti ioutil.ReadFile) balikannya
   ([]byte, error).
 - Jadi data bertipe []byte (slice of byte), err bertipe error.
 - Kenapa []byte bukan string: os.ReadFile baca file mentah
   byte-per-byte. Makanya baris 29 perlu string(data) — itu type
   conversion []byte → string.

 Asal strings.Split dan kenapa pakai range:
 - strings = package standar Go (import "strings" di config/env.go:5).
   Bukan library external.
 - strings.Split(s, sep) mengembalikan []string (slice). Di sini
   memecah isi file per baris dengan pemisah "\n".
 - range = keyword Go untuk iterasi. Karena Split hasilnya slice, mau
   dipakai satu-satu ya harus di-loop → for ... range.
 - for _, line := range → _ (blank identifier) menampung index yang
   dibuang, line menampung nilai/isi baris. Kalau ditulis for i, line
   := pun sah, cuma i nggak kepakai.

 Koreksi atas tebakanmu: Split itu untuk memisah baris, bukan
 key/value. Pemisah KEY=value pakai fungsi lain:

 ```go
key, value, found := strings.Cut(line, "=")   // env.go:34
 ```

 - strings.Cut (Go 1.18+): potong string di pemisah pertama, balikan
   (before, after, found bool).
 - found == false (baris tanpa =) → continue (lewati, lanjut iterasi
   berikutnya).
 - env.go:39 strings.Trim(..., "\"") = buang tanda petik di awal/akhir
   value. TrimSpace = buang spasi.
 - os.Getenv(key) == "" → baru os.Setenv(key, value): tidak menimpa
   env yang sudah ada di sistem.
 - strings.HasPrefix(line, "#") → baris komentar dilewati.

 Alur ringkas: baca file → []byte → string() → Split("\n") → loop per
 baris → buang spasi/komentar → Cut("=") → Setenv.

 ─────────────────────────────────────────────────────────────────────

 2. Koneksi DB — database/postgres.go:16

 ```go
db, err := database.Connect(ctx)   // db = *pgxpool.Pool
defer db.Close()
 ```

 - ctx := context.Background() (main.go:21) = context kosong sebagai
   "pembawa batas waktu/pembatalan", wajib argumen pertama untuk
   operasi I/O di Go.
 - database.Connect balikannya (*pgxpool.Pool, error) — pool koneksi
   (bukan 1 koneksi), supaya aman dipakai banyak request bersamaan
   (Fiber jalan per-goroutine).
 - log.Fatal(err) = cetak error lalu os.Exit(1) — program berhenti.
 - defer = tunda eksekusi sampai fungsi main() selesai. Jadi
   db.Close() jalan terakhir saat server mati. Pemahamanmu benar.

 ─────────────────────────────────────────────────────────────────────

 3. studentRepo := repository.StudentRepository{DB: db}

 ```go
type StudentRepository struct {   // student_repository.go:15
    DB *pgxpool.Pool
}
 ```

 - repository.StudentRepository{DB: db} = composite literal (cara
   bikin instance struct).
 - DB: db = keyed field: isi field DB dengan variabel db. Kalau tanpa
   nama → StudentRepository{db} (harus urut, tidak disarankan).
 - Asal: package latihan-fiber/app/repository (main.go:9), type-nya
   ada di app/repository/student_repository.go:15.
 - Logikanya = dependency injection manual. Struct repository itu cuma
   "cangkang"; ia butuh koneksi DB untuk menjalankan SQL. Daripada
   repository membuka koneksi sendiri (tiap repo buka koneksi sendiri
   = boros & susah dites), koneksi disuntikkan dari luar lewat field
   DB.
 - Kenapa harus ada 3 baris: karena ada 3 repository berbeda (Student,
   Prestasi, User). Tapi pool-nya satu dan sama (db). Value type
   (bukan pointer) — karena pgxpool.Pool internalnya pointer, menyalin
   struct-nya aman.

 ─────────────────────────────────────────────────────────────────────

 4. Komentar // 3. — Service (main.go:34-41)

 ```go
secret := []byte(config.GetWithDefault("JWT_SECRET",
"rahasia-latihan"))
 ```

 - []byte(...) = konversi string → []byte. Perlu karena library JWT
   minta []byte.
 - config.GetWithDefault (config/env.go:14): os.Getenv(key); kalau
   kosong pakai fallback. Jadi kalau JWT_SECRET tak ada di .env,
   dipakai "rahasia-latihan".

 ```go
svc := service.Services{
    Student:  service.StudentService{Repo: studentRepo},
    Prestasi: service.PrestasiService{Repo: prestasiRepo},
    Auth:     service.AuthService{Repo: userRepo, Secret: secret},
}
 ```

 - Lagi-lagi composite literal bersarang, field pakai nama (keyed).
 - Type Services ada di app/service/services.go:5 — struct "wadah"
   berisi 3 service.
 - StudentService (student_service.go:16) punya field Repo
   repository.StudentRepository → diisi studentRepo dari langkah 3.
 - AuthService (auth_service.go:13) punya Repo + Secret → makanya dua
   field diisi.
 - Logika: rantai db → repository → service → route. Service tidak
   tahu SQL, repository tidak tahu HTTP. svc jadi satu bungkusan yang
   dilempar ke route.

 ─────────────────────────────────────────────────────────────────────

 5. Komentar // 4. — Fiber + route (main.go:44-46)

 ```go
app := fiber.New()               // constructor dari library
github.com/gofiber/fiber/v2
app.Use(middleware.Logger)       // daftarkan middleware global
(jalan di SEMUA request)
route.Setup(app, svc, secret)    // daftarkan endpoint
 ```

 - fiber.New() = factory function dari package external
   github.com/gofiber/fiber/v2 (main.go:7), balikannya *fiber.App
   (pointer).
 - app.Use(...) = method pada *fiber.App untuk memasang middleware.
   middleware.Logger = nama fungsi dikirim tanpa tanda () — ini
   function value (Fiber memanggilnya nanti tiap ada request).
 - route.Setup(app, svc, secret) = fungsi di route/route.go:12;
   menerima *fiber.App, service.Services, []byte. Tugasnya:
   app.Group("/api/v1"), api.Get/Post/Put/..., dan
   api.Use(middleware.Protect(jwtSecret)) — baris 25 itu yang bikin
   semua route setelahnya wajib token JWT, sedangkan /auth/login
   (baris 22) masih bebas.
 - s.Student.List (tanpa kurung) juga method value: Fiber simpan
   pointer method itu sebagai handler.

 ─────────────────────────────────────────────────────────────────────

 6. Komentar // 5. — 404 (main.go:48-50)

 ```go
app.Use(func(c *fiber.Ctx) error {
    return c.Status(404).JSON(fiber.Map{"error": "endpoint tidak
ditemukan"})
})
 ```

 - func(c *fiber.Ctx) error { ... } = anonymous function / closure
   (fungsi tanpa nama, langsung ditulis di tempat) — di Fiber type ini
   bernama fiber.Handler.
 - Logika: middleware di Fiber dieksekusi sesuai urutan pendaftaran.
   Semua route sudah didaftarkan route.Setup lebih dulu (langkah 4).
   Kalau URL cocok → handler route jalan dan request selesai. Kalau
   tidak ada yang cocok, eksekusi jatuh terus ke app.Use terakhir ini
   → balas 404 JSON.
 - c.Status(404) dan .JSON(...) = method chaining: tiap method
   balikannya *fiber.Ctx lagi, jadi bisa dirantai.
 - fiber.Map{"..." : "..."} = alias buatan Fiber untuk
   map[string]interface{} — biar bisa bikin JSON object tanpa
   definisikan struct.
 - c (*fiber.Ctx) = satu object berisi semua info request + alat untuk
   balas response.

 ─────────────────────────────────────────────────────────────────────

 7. Baris terakhir (main.go:52-54)

 ```go
port := config.GetWithDefault("APP_PORT", "3000")
log.Fatal(app.Listen(":" + port))
 ```

 - ":" + port = string concatenation → ":3000" (format alamat listen
   Go).
 - app.Listen memblokir (infinite loop melayani request), balikannya
   error. log.Fatal di sini = "kalau server gagal jalan, cetak error
   lalu exit".

 ─────────────────────────────────────────────────────────────────────

 Ringkasan rantai sintaks utama

 ┌─────────────────────────────────┬───────────────────┬─────────────┐
 │ Sintaks                         │ Nama              │ Asal        │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ :=                              │ short variable    │ keyword Go  │
 │                                 │ declaration       │             │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ os.ReadFile → []byte            │ stdlib os         │ Go bawaan   │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ strings.Split/Cut/Trim/HasPrefi │ stdlib strings    │ Go bawaan   │
 │ x                               │                   │ (di-import) │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ for _, x := range               │ range loop +      │ keyword Go  │
 │                                 │ blank identifier  │             │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ Type{Field: val}                │ composite literal │ Go          │
 │                                 │ (keyed)           │             │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ []byte(s) / string(b)           │ type conversion   │ Go          │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ defer f()                       │ tunda sampai      │ keyword Go  │
 │                                 │ fungsi return     │             │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ func(...) ... {...} tanpa nama  │ anonymous         │ Go          │
 │                                 │ function/closure  │             │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ fiber.New/Use/Group/Get/Listen  │ library           │ external    │
 │                                 │ gofiber/fiber/v2  │ (go.mod)    │
 ├─────────────────────────────────┼───────────────────┼─────────────┤
 │ pgxpool.Pool                    │ library           │ external    │
 │                                 │ jackc/pgx/v5      │ (go.mod)    │
 └─────────────────────────────────┴───────────────────┴─────────────┘

 Satu koreksi penting atas tebakanmu: strings.Split memecah baris,
 strings.Cut yang memecah KEY=value. Selebihnya pemahamanmu sudah
 tepat