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
| 6 | [ISI] | | | | |
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

## D.2 Notifikasi Perubahan Kuota

[ISI: jelaskan transaksi perubahan kuota, migration 009, kunci idempotensi, dan cara worker melanjutkan pengiriman saat retry. Sisipkan SQL migration dan potongan kode penting.]

[ISI: catat hasil pengujian dengan minimal lima peserta, kegagalan sementara, dan redelivery.]

[BUKTI D2-01: tempel setelah hasil uji D.2. Tampilkan bukti kegagalan dan hitungan pengiriman per penerima setelah redelivery.]

## D.3 Bukti Pengujian

Target berikut berasal dari modul. Ganti penanda hasil setelah setiap uji.

### C.1 Jalur normal

[ISI: status lima request, empat event, empat notifikasi dengan urutan dan subjek yang sesuai, serta hitungan email.]


### C.2 Redis berhenti dan pulih

[ISI: status dan durasi tiga pendaftaran, pending saat Redis mati, dan hasil setelah pulih.]

[BUKTI C2-01: tempel setelah hasil saat Redis mati. Tampilkan respons 201 dan monitoring yang menunjukkan pending serta queue_error.]

[BUKTI C2-02: tempel setelah hasil pemulihan. Tampilkan pending kembali 0 dan total notifikasi.]

### C.3 Gagal sementara

[ISI: konfigurasi simulasi, kegagalan yang terjadi, dan hitungan dari database, API, serta log email.]

[BUKTI C3-01: tempel setelah hasil retry. Tampilkan log gagal sementara dan ketiga hitungan.]

### C.4 Gagal permanen

[ISI: status pendaftaran, perubahan archived, nilai retry, dan jumlah attempt.]

[BUKTI C4-01: tempel setelah hasil arsip. Tampilkan nilai antrean sebelum dan sesudah serta log attempt.]

### C.5 Task ganda

[ISI: ID event, jumlah pemrosesan, jumlah yang dilewati, serta jumlah email dan notifikasi.]

[BUKTI C5-01: tempel setelah hasil redelivery. Tampilkan log pemrosesan untuk ID event yang sama dan hitungan pengiriman.]

### C.6 Penghentian worker

[ISI: urutan Ctrl+C, penyelesaian pengiriman, dan penghentian worker.]

### C.7 Pemeriksaan log

[ISI: hasil pemeriksaan pesan dan field wajib pada log API dan worker.]

### Percobaan pembanding

[ISI: bug yang dikembalikan, hasil uji saat bug aktif, dan hasil uji yang sama setelah perbaikan dipulihkan.]

[BUKTI D3-01: tempel setelah hasil pembanding. Tampilkan keluaran gagal dan lulus pada uji yang sama.]

## D.4 Analisis

1. [ISI: dua skenario ketika mencoba enqueue ulang di handler masih bisa menghilangkan notifikasi atau menunda respons.]
2. [ISI: dampak Redis kehilangan 100 task pending, cara mendeteksi dari PostgreSQL, dan cara memulihkan tanpa pengiriman ganda.]
3. [ISI: alasan pemakaian WithTx di service KRS dan WithTxOnce di worker, serta dampak jika ditukar.]
4. [ISI: kapan urutan email menjadi masalah, satu cara mengatasinya, dan konsekuensinya.]

## Kesimpulan

[ISI: rangkum perilaku yang terbukti dari pengujian.]

## Daftar pustaka

Program Studi D4 Teknik Informatika. (t.t.). *Pertemuan 11 — Background Jobs & Asynchronous Processing* [Modul praktikum]. Fakultas Vokasi, Universitas Airlangga.
