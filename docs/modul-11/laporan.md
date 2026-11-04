# Laporan Praktikum Modul 11

**Background Jobs dan Asynchronous Processing**

Identitas dan tautan repository dilengkapi pada cover draf Word.

> Panduan pengisian untuk agent. Setelah satu uji atau temuan terverifikasi, ganti penanda di bagian terkait dengan narasi siap salin ke Word. Tulis tindakan, hasil aktual, dan apa yang dibuktikan dalam 2–4 kalimat. Jika screenshot diperlukan, minta sebelum kondisi uji berubah dan catat ID bukti, isi yang harus terlihat, caption, serta lokasi tempel. Simpan hasil yang belum tersedia sebagai `[ISI]`. Progres kerja dicatat di `rencana.md`.

## Pendahuluan

Praktikum ini memproses notifikasi KRS di belakang layar. API mencatat perubahan KRS dan event ke PostgreSQL dalam satu transaksi. Relay memindahkan event dari tabel outbox ke antrean Redis. Worker mengambil task dan menjalankan pengiriman email tiruan.

**Lingkungan praktikum** — [ISI: baseline proyek dan versi alat yang benar-benar dipakai.]

## D.1 Berita Acara Perbaikan

Isi satu baris untuk tiap kesalahan bawaan modul. Untuk kesalahan yang lolos compiler, rujuk bukti sebelum dan sesudah perbaikan.

| No. | Lokasi (file dan fungsi) | Gejala | Uji | Akar masalah | Perbaikan |
|---|---|---|---|---|---|
| 1 | `helper/mailer.go`, `NewLogMailer` | Build gagal | `go build ./app/... ./helper/...` | Nilai `LogMailer` tidak memenuhi interface karena `Send` memakai penerima pointer | Mengembalikan `&LogMailer` |
| 2 | `app/repository/course_repository.go`, `Enroll` (Langkah 6 PDF) | Pendaftaran duplikat ditolak, tetapi jumlah event naik dari 1 ke 2 | POST pendaftaran IF101 dua kali dan kueri jumlah event | INSERT outbox memakai pool di luar transaksi KRS | Memanggil `outbox.WithTx(tx).Insert(...)` |
| 3 | `app/repository/course_repository.go`, `Drop` (Langkah 6 PDF) | Keluar kelas berhasil, tetapi tercatat sebagai `enrollment.created` | DELETE pendaftaran IF101 dan kueri jenis event terbaru | `Drop` memakai konstanta untuk pendaftaran | Mengganti konstanta menjadi `model.EventEnrollmentDropped` |
| 4 | `app/service/outbox_relay.go`, `publish` (Langkah 7 PDF) | Build gagal pada pemanggilan `asynq.NewTask` | `go build` | Payload berupa struct `NotificationJob`, sedangkan Asynq menerima `[]byte` | Mengubah payload menjadi JSON dengan `json.Marshal` dan menangani error serialisasi |
| 5 | `app/service/notification_worker.go`, `HandleSend` (Langkah 8 PDF) | Build gagal saat handler didaftarkan di `worker.go` | `go build .` | Parameter handler berupa `asynq.Task`, sedangkan `mux.HandleFunc` mengharapkan `*asynq.Task` | Mengubah tipe parameter menjadi `*asynq.Task` |
| 6 | `config/queue.go`, `NewWorkerServer` (Langkah 9 PDF) | Pendaftaran berhasil, tetapi daftar notifikasi kosong dan satu task menunggu di Redis | `GET /api/v1/notifications/mine` sebelum dan sesudah worker dimulai ulang | Worker mendengarkan `notification:send`, sedangkan relay mengirim ke antrean `notifications` | Mengatur worker agar mendengarkan `model.QueueNotifications` |
| 7 | [ISI] | | | | |
| 8 | [ISI] | | | | |
| 9 | [ISI] | | | | |

### 2.1 Bukti temuan

**Temuan 1 — Mailer tiruan gagal dikompilasi**

Pada kondisi awal, `NewLogMailer` mengembalikan nilai `LogMailer{...}`. Perintah `go build ./app/... ./helper/...` gagal karena metode `Send` didefinisikan pada pointer `*LogMailer`. Konstruktor diperbaiki agar mengembalikan `&LogMailer{...}`. Setelah itu, perintah yang sama selesai tanpa error.

**[D1-01 — Tempel di Word, BAB II → 2.1, setelah narasi Temuan 1. Tampilkan perintah build dan pesan compiler. Caption: Compiler menolak nilai LogMailer yang dikembalikan oleh NewLogMailer.]**

Salinan keluaran compiler tersimpan di `bukti-mailer-gagal.txt`.

**Temuan 2 — Pendaftaran duplikat meninggalkan event**

Untuk akun uji ID 6, kueri menghitung event `enrollment.created` dari 0 sebelum pendaftaran menjadi 1 setelah `POST /api/v1/courses/1/enroll` berhasil (HTTP 201). Permintaan yang sama kemudian ditolak sebagai duplikat (HTTP 409), tetapi jumlah event naik menjadi 2. Penyebabnya adalah INSERT outbox berjalan lewat pool, sehingga pembatalan transaksi KRS tidak membatalkan event tersebut.

**[D1-02 — Tempel di Word, BAB II → 2.1, setelah kondisi awal Temuan 2. Urutan caption: Pendaftaran pertama ke IF101 berhasil (HTTP 201). → Satu event tercatat setelah pendaftaran pertama. → Pendaftaran ulang ditolak (HTTP 409). → Jumlah event naik dari 1 ke 2 meski pendaftaran duplikat ditolak.]**

Pemanggilan `outbox.Insert(...)` diganti dengan `outbox.WithTx(tx).Insert(...)` agar kedua penulisan ikut transaksi yang sama. Setelah API dimulai ulang, permintaan duplikat dan kueri yang sama diulang. Jumlah tetap 2 karena percobaan ulang tidak menambah event; dua baris dari pengujian sebelumnya masih tersimpan.

**[D1-03 — Tempel setelah hasil uji ulang Temuan 2. Tampilkan kueri dan hasil 2. Caption: Setelah perbaikan, percobaan duplikat tidak menambah event outbox (jumlah tetap 2).]**

**Temuan 3 — Keluar kelas tercatat sebagai pendaftaran**

Permintaan `DELETE /api/v1/courses/1/enroll` berhasil (HTTP 204), tetapi kueri event terbaru untuk akun ID 6 dan kelas ID 1 menunjukkan `4 | enrollment.created`. Fungsi `Drop` memakai konstanta `EventEnrollmentCreated`. Konstanta itu diganti dengan `EventEnrollmentDropped`. Setelah API dimulai ulang dan akun mengambil kelas yang sama lagi, permintaan keluar diulang (HTTP 204) dan kueri yang sama menunjukkan `6 | enrollment.dropped`.

**[D1-04 — Tempel di Word, BAB II → 2.1, setelah narasi Temuan 3. Tampilkan dua hasil kueri. Caption sebelum: Keluar dari IF101 berhasil, tetapi outbox mencatat enrollment.created. Caption sesudah: Setelah perbaikan, keluar dari IF101 mencatat enrollment.dropped.]**

**Temuan 4 — Payload relay gagal dikompilasi**

Perintah `go build` gagal karena `asynq.NewTask` menerima payload `[]byte`, sementara relay memberikan struct `model.NotificationJob`. Fungsi `publish` diperbaiki dengan mengubah struct menjadi JSON menggunakan `json.Marshal` sebelum membuat task. Setelah perbaikan, perintah `go build` yang sama berhasil.

**[D1-05 — Tempel di Word, BAB II → 2.1, setelah narasi Temuan 4. Tampilkan perintah `go build` dan pesan compiler. Caption: Compiler menolak payload `NotificationJob` pada `asynq.NewTask`.]**

**[D1-06 — Tempel setelah D1-05. Tampilkan perintah `go build` dan prompt yang kembali tanpa error. Caption: Build berhasil setelah payload `NotificationJob` diubah menjadi JSON.]**

**Temuan 5 — Tanda tangan handler worker gagal dikompilasi**

Perintah `go build .` gagal saat `worker.HandleSend` didaftarkan ke `mux.HandleFunc` di `worker.go`. Parameter `task` pada `HandleSend` bertipe `asynq.Task`, sedangkan Asynq mengharapkan `*asynq.Task`. Setelah tipe parameter diganti menjadi `*asynq.Task`, perintah `go build .` yang sama berhasil.

**[D1-07 — Tempel di Word, BAB II → 2.1, setelah narasi Temuan 5. Tampilkan perintah `go build .` dan pesan compiler. Caption: Compiler menolak tanda tangan `HandleSend` saat didaftarkan ke Asynq.]**

**[D1-08 — Tempel setelah D1-07. Tampilkan perintah `go build .` dan prompt yang kembali tanpa error. Caption: Build berhasil setelah parameter `HandleSend` memakai `*asynq.Task`.]**

**Temuan 6 — Worker mendengarkan antrean yang keliru**

Setelah akun `mhstahap2` mendaftar ke IF102, `GET /api/v1/notifications/mine` mengembalikan HTTP 200 dengan isi `[]`. Pemeriksaan Redis menunjukkan satu task menunggu di antrean `notifications`, sementara worker dikonfigurasi untuk mendengarkan `notification:send`, yaitu jenis task. Konfigurasi diganti ke `model.QueueNotifications` dan worker dimulai ulang. Request GET yang sama kemudian mengembalikan satu notifikasi dengan `outbox_id: 7` tanpa mengulang pendaftaran.

**[D1-09 — Tempel di Word, BAB II → 2.1, setelah narasi Temuan 6. Tampilkan GET, URL, HTTP 200, dan respons kosong. Caption: Pendaftaran IF102 berhasil, tetapi daftar notifikasi akun uji masih kosong.]**

**[D1-10 — Tempel setelah D1-09. Tampilkan GET yang sama, HTTP 200, dan satu notifikasi dengan outbox_id 7. Caption: Setelah nama antrean diperbaiki, worker memproses event IF102.]**

## D.2 Notifikasi Perubahan Kuota

Migration 009 mengganti kunci unik menjadi `(outbox_id, user_id)`. Update kuota, snapshot ID peserta, dan insert event `course.quota_changed` berjalan dalam satu transaksi. Worker memakai transaksi claim terpisah untuk setiap penerima. Gagal sementara membatalkan claim penerima tersebut sehingga retry dapat melanjutkan pengiriman dan melewati penerima yang telah berhasil.

PATCH kelas 7 menghasilkan HTTP 200 dengan kuota 41 dan event ID 36 untuk lima peserta. Terjadi dua kegagalan sementara. Setelah retry dan satu redelivery, terdapat satu baris notifikasi dan satu baris email per penerima. Log redelivery menampilkan lima `already_sent=true`. Keluaran tersimpan di `logs/speed-quota.log` dan `logs/mail.log`.

## D.3 Bukti Pengujian

Target berikut berasal dari modul. Ganti penanda hasil setelah setiap uji.

### Uji jalur dasar Tahap 2

API dan worker dijalankan di dua proses. Redis lokal merespons `PONG`. Worker memakai `RELAY_AFTER_ID=6` pada database `praktikum_backend`, sehingga lima event pending Tahap 1 berada di luar cakupan relay selama uji. Satu request `POST /api/v1/courses/2/enroll` untuk akun `mhstahap2` (ID 7) menghasilkan pendaftaran IF102. Setelah perbaikan antrean, API menampilkan satu notifikasi email dengan `outbox_id: 7`, dan pencarian berdasarkan penerima serta subjek di `logs/mail.log` menemukan tepat satu baris.

**[T2-01 — Tempel di Word pada D.3 → Uji jalur dasar Tahap 2. Tampilkan POST pendaftaran dan respons yang memuat IF102. Caption: Akun uji berhasil terdaftar di IF102.]**

**[T2-02 — Tempel setelah T2-01. Tampilkan perintah Select-String dan satu baris yang cocok di logs/mail.log. Caption: Pendaftaran IF102 menghasilkan satu email tiruan untuk akun uji.]**

### C.1 Jalur normal

Permintaan pindah dari kelas 2 (IF102) ke kelas 4 (IF302) menghasilkan HTTP 200 dan mengembalikan kelas tujuan IF302 Keamanan Informasi. Saat diagnosis, port 3000 mengembalikan halaman HTML JIA Admin. Bukti swap berhasil diambil setelah API praktikum dijalankan pada port 3001.

**[C1-01e — Tempel di Word pada D.3 → C.1 Jalur normal. Tampilkan POST swap, body dengan kelas 2 dan 4, HTTP 200, serta kode IF302. Caption: API berhasil memindahkan KRS dari IF102 ke IF302.]**

Kueri outbox untuk `mhsuji` menghasilkan empat event berurutan: `enrollment.created` (ID 1), `enrollment.dropped` (ID 3), `enrollment.created` (ID 4), dan `enrollment.swapped` (ID 5). Log email berisi tepat empat baris untuk `mhsuji@student.unair.ac.id`.

**[C1-02b — Tempel di Word pada D.3 → C.1 Jalur normal. Tampilkan kueri dengan empat baris outbox serta perintah hitung email dengan hasil 4. Caption: Empat event KRS tercatat berurutan dan menghasilkan empat email tiruan.]**

Screenshot Word mengonfirmasi status 201, 409 ALREADY_ENROLLED, 204, 201, dan 200. GET menampilkan empat notifikasi. Tiga objek terlipat pada screenshot sehingga ketiga subjek tersebut belum bisa diverifikasi dari gambar.


### C.2 Redis berhenti dan pulih

Saat Redis mati, pendaftaran kelas 5, 6, dan 7 masing-masing berhasil dengan HTTP 201. Monitoring menampilkan `queue_error`, tetapi `outbox_pending` bernilai 0 (target 3), `outbox_published` naik dari 4 ke 7, dan `notifications_sent` tetap 4.

**[Bukti sebelum perbaikan — Tempel di Word pada D.3 → C.2. Caption: Saat Redis mati, tiga pendaftaran berhasil tetapi event langsung ditandai terbit.]**

**Temuan 7 untuk D.1 — Event ditandai terbit saat Redis gagal menerima task.** Di `OutboxRelay.RelayOnce`, ID event semula ditambahkan ke daftar `MarkPublished` sebelum upaya enqueue. Pemanggilan `append` dipindah setelah pengecekan hasil enqueue, sehingga event yang gagal tetap menunggu dan bisa dicoba kembali.

**[Bukti diff — Tempel di D.1 pada Temuan 7. Caption: ID event baru masuk ke daftar penandaan terbit setelah enqueue berhasil.]**

Setelah perbaikan, worker dijalankan ulang dan status tiga event yang terdampak (ID 6–8) dikembalikan menjadi menunggu. Selama Redis mati, monitoring menunjukkan `outbox_pending: 3`, `outbox_published: 4`, dan `queue_error`. Setelah Redis dinyalakan, pending turun menjadi 0 dan total notifikasi terkirim naik dari 4 menjadi 7. Durasi tiga request dan waktu pemulihan belum terukur.

**[Bukti setelah perbaikan — Tempel di D.3 → C.2. Caption: Tiga event tetap menunggu saat Redis mati dan seluruhnya diproses setelah Redis menyala kembali.]**

Pengguna mengonfirmasi bukti tujuh notifikasi akun `mhsuji` telah disimpan sebelum reset berikutnya.

### C.3 Gagal sementara

Pada uji `MAILER_FAIL_EVERY=3`, database mencatat 6 notifikasi, sedangkan log berisi 4 email untuk `mhsuji`. Selisih dua menunjukkan catatan notifikasi tersimpan meskipun pengiriman gagal.

**Temuan 8 untuk D.1 — Claim notifikasi berada di luar transaksi.** `NotificationWorker.HandleSend` memulai transaksi, tetapi `Claim` semula memakai koneksi repository biasa. Perbaikan memakai `WithTx(tx)` agar pengembalian error dari mailer membatalkan claim dan retry dapat mengirim ulang.

**[Bukti sebelum perbaikan — Tempel di D.3 pada uji gagal sementara. Caption: Database mencatat enam notifikasi, tetapi log hanya berisi empat email.]**

**[Bukti diff — Tempel di D.1 pada Temuan 8. Caption: Claim notifikasi menggunakan transaksi pengiriman.]**

Setelah perbaikan, API menampilkan enam notifikasi. Log worker mencatat dua kegagalan sementara dan dua pemrosesan ulang yang berhasil dengan `retry=1`. Pemeriksaan log email menunjukkan enam baris untuk akun uji.

### C.4 Gagal permanen

Pendaftaran akun `salahalamat` ke kelas 2 menghasilkan HTTP 201. Task event 28 sempat masuk retry untuk alamat `salah@kampus.invalid`, padahal targetnya langsung diarsipkan. Kesalahan permanen perlu membungkus sentinel `asynq.SkipRetry` dengan `%w` agar Asynq mengenalinya.

Setelah perbaikan, akun uji kedua dengan alamat `ulang@kampus.invalid` mendapat HTTP 201. Archived naik dari 4 ke 5 dan retry tetap 0. Log mencatat tepat satu `task_failed` pada attempt 1/6.

### C.5 Task ganda

Probe redelivery mengirim ulang satu salinan untuk outbox ID 15. Task masuk retry dengan pesan `claim notifikasi berhasil dibuat`. Cabang untuk claim yang sudah ada perlu selesai dengan sukses sehingga worker mencatat `already_sent=true` dan melewati pengiriman.

Setelah perbaikan, lima salinan dibuat untuk event 30 saat worker berhenti, lalu relay menambah task asli. Log menunjukkan enam pemrosesan, satu `already_sent=false` dan lima `true`. Database dan log email masing-masing mencatat satu pengiriman.

### C.6 Penghentian worker

Pendaftaran kelas 9 menghasilkan HTTP 201. Uji otomatis mengirim CTRL_BREAK saat antrean menunjukkan satu task aktif dan mailer memakai jeda lima detik. Worker mencatat `notification_processed` untuk event 37 sebelum keluar dengan kode 0. API menampilkan notifikasi IF307 Sistem Terdistribusi.

### C.7 Pemeriksaan log

Log worker menunjukkan `relay_done published`, `relay_delayed published/error`, `notification_processed outbox_id/event/retry/already_sent`, dan `task_failed type/attempt/error`. Tidak ada `request_failed` pada log API yang diarsipkan dari pengujian otomatis. Log API interaktif awal tidak diarsipkan secara lengkap.

### Percobaan pembanding

Percobaan pembanding memakai dua binary. Binary sementara dibuat dengan Go overlay yang mengembalikan error pada cabang task duplikat. Redelivery event 30 mencatat `task_failed`. Uji yang sama pada binary perbaikan mencatat `notification_processed already_sent=true`. Build, vet, dan test seluruh paket berhasil.

## D.4 Analisis

1. Proses API dapat berhenti sesudah transaksi KRS berhasil tetapi sebelum antrean menerima task. Jika Redis mati lebih lama daripada tiga percobaan, semua percobaan habis dan respons pengguna tertunda oleh timeout. Outbox menyimpan pekerjaan bersama transaksi KRS agar proses lain bisa mencoba kembali.
2. Jika Redis kehilangan 100 task pending, outbox dapat tetap berstatus terbit walaupun task tidak akan dijalankan. Bandingkan event lama yang sudah terbit dengan catatan notifikasi per penerima, lalu singkirkan task yang masih aktif dan kegagalan permanen dari daftar pemulihan. Kirim ulang event yang belum lengkap. Claim unik menahan pengiriman ganda untuk penerima yang sudah tercatat.
3. Helper transaksi yang dapat mengulang transaksi cocok untuk KRS selama isi transaksinya hanya operasi PostgreSQL yang dibatalkan bersama. Worker perlu satu percobaan transaksi karena pengiriman email punya efek di luar database. Mengulang transaksi worker secara otomatis dapat mengirim email dua kali. Dengan satu percobaan, kegagalan KRS yang sebenarnya bisa dicoba kembali akan diteruskan ke pemanggil. Repo ini memakai `pgx.BeginFunc` untuk satu percobaan.
4. Email pendaftaran dan keluar dapat tiba terbalik ketika worker berjalan paralel atau task pertama mengalami retry. Jika urutan penting, gunakan nomor urut per mahasiswa dan proses satu event per mahasiswa pada satu waktu. Task yang gagal dapat menahan pesan sesudahnya dan memperlambat pengiriman.

## Kesimpulan

Outbox menjaga event KRS saat Redis tidak tersedia. Worker mencoba kembali kegagalan sementara, langsung mengarsipkan alamat invalid, dan mengabaikan task duplikat yang sudah diproses. Perubahan kuota untuk lima peserta menghasilkan satu email per penerima walaupun terjadi retry dan redelivery.

## Daftar pustaka

Program Studi D4 Teknik Informatika. (t.t.). *Pertemuan 11 — Background Jobs & Asynchronous Processing* [Modul praktikum]. Fakultas Vokasi, Universitas Airlangga.
