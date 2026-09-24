 BAGIAN A — Cara pembuatannya (urutan dari bawah ke atas)

 A1. Tabel dulu — migrations/002_create_prestasi.sql

 ```sql

CREATE TABLE IF NOT EXISTS prestasi (

    id SERIAL PRIMARY KEY,

    student_id INTEGER NOT NULL,

    keterangan VARCHAR(100) NOT NULL,

    juara INTEGER NOT NULL CHECK (juara &gt;= 0 AND juara &lt;= 100),

    ...

    CONSTRAINT fk_prestasi_student FOREIGN KEY (student_id)

        REFERENCES students (id) ON DELETE CASCADE

);

 ```

 - SERIAL = auto-increment (setara INTEGER + sequence), asalnya

   PostgreSQL.

 - CHECK (juara &gt;= 0 AND juara &lt;= 100) = validasi di level database.

 - ON DELETE CASCADE = mahasiswa dihapus → prestasinya ikut terhapus.

 A2. Model — app/model/prestasi.go:4

 ```go

type Prestasi struct {

    ID         int    `json:"id"`

    StudentID  int    `json:"student_id"`

    Keterangan string `json:"keterangan"`

    Juara      int    `json:"juara"`

}

 ```

 - json:"..." = struct tag (metadata). Dipakai encoding/json saat

   c.JSON() untuk menentukan nama key di JSON.

 - Kenapa perlu struct: SQL mengembalikan baris mentah; struct ini

   "cetakan" untuk menampungnya.

 A3. Repository — app/repository/prestasi_repository.go

 Struct penampung pool (prestasi_repository.go:15):

 ```go

type PrestasiRepository struct { DB *pgxpool.Pool }

 ```

 Fungsi intinya (:102):

 ```go

func (r PrestasiRepository) GetByStudentID(ctx context.Context,

studentID int) ([]model.Prestasi, error)

 ```

 - (r PrestasiRepository) sebelum nama fungsi = receiver / method.

   Jadi ini bukan fungsi bebas, melainkan method milik struct →

   dipanggil repo.GetByStudentID(...), dan r.DB bisa diakses di

   dalamnya.

 - Balikan 3 nilai dibaca: []model.Prestasi (slice hasil), error.

 A4. Service — app/service/prestasi_service.go:60

 ```go

func (s PrestasiService) GetByStudent(c *fiber.Ctx) error

 ```

 - c *fiber.Ctx = tanda ini adalah HTTP handler (Fiber yang memanggil,

   bukan kita).

 - Service tidak sentuh SQL — ia panggil s.Repo.GetByStudentID(...).

 A5. Route — route/route.go:36

 ```go

st.Get("/:id/prestasi", s.Prestasi.GetByStudent)

 ```

 - st = [api.Group](http://api.Group)("/students"), api = [app.Group](http://app.Group)("/api/v1") → path

   penuh /api/v1/students/:id/prestasi.

 - :id = path parameter Fiber.

 - s.Prestasi.GetByStudent tanpa kurung = method value: alamat method

   dikirim, bukan dipanggil.

 A6. Wiring di main.go

 ```

prestasiRepo := repository.PrestasiRepository{DB: db}   // baris 31

Prestasi: service.PrestasiService{Repo: prestasiRepo},  // baris 39

route.Setup(app, svc, secret)                            // baris 46

 ```

 Rantai lengkapnya: db → PrestasiRepository → PrestasiService →

 Services → route.Setup → endpoint hidup.

 ─────────────────────────────────────────────────────────────────────

 BAGIAN B — Yang terjadi saat dipanggil dari Postman

 Misal: GET [http://localhost:3000/api/v1/students/1/prestasi](http://localhost:3000/api/v1/students/1/prestasi)

 B0. Syarat: token dulu

 Di route.go:25 ada api.Use(middleware.Protect(jwtSecret)). Endpoint

 prestasi terdaftar sesudah baris itu (baris 36) → wajib token.

 1. POST /api/v1/auth/login → dapat token.

 2. Di tab Authorization → type Bearer Token → tempel token.

 Kalau lupa: middleware/auth.go:19 cek strings.HasPrefix(header,

 "Bearer ") → gagal → 401 "token belum dikirim...".

 B1. Fiber cocokkan URL

 Fiber scan route yang terdaftar. /api/v1/students/1/prestasi cocok

 dengan pola /:id/prestasi → :id = "1" (masih string).

 B2. Protect lolos → [c.Next](http://c.Next)()

 middleware/auth.go:34 [c.Next](http://c.Next)() = "lanjut ke handler berikutnya".

 Sambil menyimpan identitas: c.Locals("user_id", claims["sub"]).

 B3. Service jalan — prestasi_service.go:61

 ```go

id, err := strconv.Atoi(c.Params("id"))

 ```

 - c.Params("id") = method Fiber mengambil path param → string "1".

 - strconv.Atoi = fungsi stdlib Go (strconv = string conversion), Atoi

   = ASCII to int. Balikan (int, error).

 - err != nil || id &lt; 1 → 400. Jadi /students/abc/prestasi atau

   /students/0/prestasi ditolak di sini, sebelum menyentuh database.

 B4. Panggil repository — :65

 ```go

items, err := s.Repo.GetByStudentID(c.Context(), id)

 ```

 - c.Context() = method Fiber mengambil context.Context milik request

   (pembawa pembatalan/batas waktu).

 - Gagal → helper.Error(c, 500, ...) → JSON {"error": "..."}.

 B5. Eksekusi SQL — prestasi_repository.go:103

 ```go

rows, err := r.DB.Query(ctx,

    `SELECT id, student_id, keterangan, juara FROM prestasi WHERE

student_id = $1 ORDER BY id`,

    studentID,

)

 ```

 - Backtick ` = raw string literal (bisa multi-baris, \n tidak

   diproses) — enak untuk SQL panjang.

 - $1 = placeholder pgx/pg (bukan Fiber, bukan Go). Nilainya dari

   argumen berikutnya (studentID). Kenapa tidak di-concat: mencegah

   SQL injection dan tipe data ditangani driver.

 - r.DB.Query → rows bertipe pgx.Rows (cursor ke hasil query).

 ```go

defer rows.Close()   // baris 110

 ```

 - defer = jalankan rows.Close() saat fungsi return — koneksi          

   dikembalikan ke pool. Wajib, kalau tidak pool bocor.                

                                                                       

 B6. Baris SQL → struct — :112-119                                     

                                                                       

 ```go                                                                 

out := []model.Prestasi{}      // slice kosong, BUKAN nil

for [rows.Next](http://rows.Next)() {              // maju ke baris berikutnya; false

bila habis

    var p model.Prestasi

    if err := rows.Scan(&amp;[p.ID](http://p.ID), &amp;p.StudentID, &amp;p.Keterangan,

&amp;p.Juara); err != nil { ... }

    out = append(out, p)

}

return out, rows.Err()

 ```                                                                   

                                                                       

 - out := []model.Prestasi{} = composite literal slice kosong.         

   Penting: kalau pakai var out []model.Prestasi (nil slice), JSON-nya 

   jadi null; dengan {} jadi [].                                       

 - [rows.Next](http://rows.Next)() = iterator; tiap loop satu baris.                       

 - rows.Scan(&amp;[p.ID](http://p.ID), ...) = tanda &amp; = alamat memori (pointer). Scan     

   butuh pointer supaya bisa menulis ke field p. Kalau tanpa &amp;, nilai  

   hanya disalin dan p tetap kosong.                                   

 - append(out, p) = stdlib builtin menambah elemen ke slice.           

 - rows.Err() = error yang terjadi saat iterasi (beda dengan error     

   awal Query).                                                        

                                                                       

 B7. Balas — prestasi_service.go:69                                    

                                                                       

 ```go                                                                 

return c.JSON(items)

 ```                                                                   

                                                                       

 - c.JSON = method Fiber → set header Content-Type: application/json + 

   kirim.                                                              

 - Struct tag json:"student_id" di model yang membuat key JSON bernama 

   student_id, bukan StudentID.                                        

                                                                       

 ───────────────────────────────────────────────────────────────────── 

                                                                       

 BAGIAN C — Ringkasan alur                                             

                                                                       

 ```                                                                   

Postman GET /api/v1/students/1/prestasi  (+ header Authorization:

Bearer &lt;token&gt;)

   ↓

Fiber cocokkan route (route.go:36)

   ↓

middleware.Logger (main.go:45) → catat request

   ↓

middleware.Protect (route.go:25 → auth.go:15) → cek token → [c.Next](http://c.Next)()

   ↓

PrestasiService.GetByStudent (prestasi_service.go:60)

     strconv.Atoi(c.Params("id"))          → "1" jadi 1, validasi &gt; 0

   ↓

PrestasiRepository.GetByStudentID (prestasi_repository.go:102)

     DB.Query(SQL, $1)  → rows

     defer rows.Close()

     for [rows.Next](http://rows.Next)() { rows.Scan(&amp;p...) ; append }

   ↓

c.JSON(items) → 200 OK

 ```

 ─────────────────────────────────────────────────────────────────────

 Tabel sintaks baru di endpoint ini

 ┌──────────────────────────┬───────────────────────┬────────────────┐

 │ Sintaks                  │ Nama                  │ Asal           │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ `...`                    │ raw string literal    │ Go             │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ func (r T) Nama()        │ method + receiver     │ Go             │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ &amp;[p.ID](http://p.ID)                    │ address-of / pointer  │ Go             │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ append(s, v)             │ builtin tambah elemen │ Go             │

 │                          │ slice                 │                │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ []model.Prestasi{}       │ composite literal     │ Go             │

 │                          │ slice kosong          │                │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ strconv.Atoi             │ string → int          │ stdlib strconv │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ fmt.Errorf(... %w)       │ bungkus error         │ stdlib fmt     │

 │                          │ (wrapping)            │                │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ c.Params("id")           │ ambil path param      │ Fiber          │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ c.QueryInt(k, default)   │ ambil query string    │ Fiber          │

 │                          │ ?page=                │                │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ c.Context()              │ ambil context.Context │ Fiber          │

 │                          │ request               │                │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ c.JSON(v) /              │ kirim JSON / simpan   │ Fiber          │

 │ c.Locals(k,v) / [c.Next](http://c.Next)() │ data / lanjut         │                │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ rows.Next/Scan/Err       │ iterator hasil query  │ jackc/pgx/v5   │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ $1                       │ placeholder parameter │ PostgreSQL/pgx │

 │                          │ SQL                   │                │

 ├──────────────────────────┼───────────────────────┼────────────────┤

 │ ON DELETE CASCADE, CHECK │ 约束 tabel            │ PostgreSQL     │

 └──────────────────────────┴───────────────────────┴────────────────┘

 Satu catatan praktis: endpoint ini tidak punya pagination (beda

 dengan List di prestasi_service.go:24 yang baca ?page/?limit).

 GetByStudentID mengambil semua prestasi mahasiswa itu sekaligus. 