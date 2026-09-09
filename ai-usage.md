# Penggunaan AI dalam Tugas Ini

Dokumen ini menjelaskan peran kecerdasan buatan (AI) pada setiap tugas, serta
prinsip umum yang saya ikuti selama penggunaannya.

## Tugas 1

> Tugas ini saya kerjakan sendiri. Beberapa hal seperti penjelasan konsep-konsep
> saya minta untuk dijelaskan ulang oleh Claude.

Pada tugas ini AI (Claude) hanya berperan sebagai **asisten penjelasan konsep**.
Saya mempraktikkan konsep tersebut di kode, kemudian meminta Claude untuk
menjelaskan ulang konsep-konsep yang belum saya pahami secara mendalam.

## Tugas 2

> Tugas ini saya kerjakan sendiri. Beberapa hal seperti penjelasan konsep-konsep
> saya minta untuk dijelaskan ulang oleh Gemini dengan alur sebagai berikut:
> minta Gemini untuk menerangkan rangkuman dari laporan yang diberikan, lanjutkan
> tonton sumber belajar dari Google dan YouTube, lalu lanjutkan membuat sendiri
> kode sederhana terlebih dahulu. Jika berhasil, lanjutkan. Jika tidak, tanya ke AI
> untuk troubleshooting dan alternatif penyelesaian, kemudian revisi dan evaluasi
> apa saja yang bisa ditingkatkan.

Alur kerja AI (Gemini) :

1. **Rangkuman laporan** — meminta rangkuman dari laporan/Modul yang diberikan.
2. **Sumber belajar** — mencari sumber belajar (Google, YouTube) untuk konsep
   yang relevan.
3. **Implementasi** — mempraktikkan kode sederhana berdasarkan pemahaman
   yang didapat dari diskusi dengan AI.
4. **Troubleshooting** — bila kode belum berjalan, meminta AI untuk mengidentifikasi
   masalah dan memberikan alternatif penyelesaian.
5. **Revisi & evaluasi** — memperbaiki kode, kemudian mengevaluasi apa saja yang
   bisa ditingkatkan.

## Tugas 3

> ChatGPT digunakan untuk membantu mengidentifikasi modul, menganalisis modul,
> dan memberikan langkah-langkah pengerjaan. Bantuan difokuskan pada
> langkah-langkah pengerjaan dan penyusunan outline laporan, parafrair bahasa,
> dan penempatan placeholder screenshot.

AI (ChatGPT) membantu pada:

- **Identifikasi & analisis modul** — menentukan modul apa saja yang relevan.
- **Langkah pengerjaan** — merancang urutan langkah-langkah kerja.
- **Outline laporan** — menyusun struktur laporan.
- **Parafrair bahasa** — menyederhanakan atau memperbaiki bahasa agar lebih jelas.
- **Placeholder screenshot** — menata letak screenshot dalam laporan.

## Tugas 4

> AI (ChatGPT) digunakan sebagai alat bantu dalam membaca modul, brainstorm, dan
> memberi gambaran besar hingga ke detail implementasinya dengan cara terus
> bertanya dan melakukan iterasi. Setelah mendapatkan langkah demi langkah
> secara spesifik, saya mengerjakan berdasarkan urutan tersebut, melakukan
> troubleshooting hingga selesai dan diverifikasi dengan `go test`, `go vet`,
> `go build`, pemeriksaan dependency dengan `go list`, serta pengujian endpoint
> langsung melalui Postman lokal.

Alur kerja AI (ChatGPT) :

1. **Baca modul & brainstorming** — membaca materi dan berdiskusi untuk
   merancang gambaran besar (high-level) hingga detail (low-level).
2. **Iterasi bertanya** — terus bertanya untuk memperhalus pemahaman.
3. **Step-by-step outline** — meminta AI membuat alur langkah-demi-langkah
   pengerjaan termasuk outline laporan dan urutan refactoring.
4. **Implementasi** — mempraktikkan kode berdasarkan langkah yang didapat
   dari hasil diskusi dengan AI.
5. **Troubleshooting & verifikasi** — memperbaiki kode hingga stabil, kemudian
   memverifikasi dengan:
   - `go test` — menjalankan unit/integrasi tes.
   - `go vet` — memeriksa potensi bug statis.
   - `go build` — memastikan kode dapat dikompilasi.
   - `go list` — memeriksaan dependency.
   - **Postman lokal** — uji endpoint secara langsung.

---

## Prinsip Umum Penggunaan AI (General)

> AI digunakan untuk **identifikasi**, **analisis**, dan **menemukan alternatif
> solusi** yang relevan. Saya ingin AI fokus pada brainstorming dan memanfaatkan
> wawasan yang kurang saya miliki. Namun, untuk praktiknya, saya tetap
> mempraktikkan kode berdasarkan hasil diskusi dengan AI, bukan menyalin
> langsung. AI bukan sebagai agen atau subjek pertama, melainkan sebagai
> **pihak ketiga** yang membantu percepatan pemecahan dan analisis masalah
> serta solusi yang tersedia.

Ringkasnya:

- **AI = asisten analisis & brainstorming**, bukan pengganti praktik saya.
- **AI membantu identifikasi** akar permasalahan (identifikasi).
- **AI membantu analisis** dampak yang terdampak (analisis).
- **AI memberi alternatif** solusi yang dapat saya terapkan (solusi).
- **Saya yang menerapkan** keputusan akhir ke dalam kode secara mandiri,
  berdasarkan hasil diskusi dengan AI.

AI berperan sebagai **"second pair of eyes"** — memberi perspektif ekstra, namun
keputusan akhir dan implementasi selalu ada di tangan saya.
