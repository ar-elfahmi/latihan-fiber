-- ============================================================
-- test_setup.sql — Reset dan seed database untuk batch testing
-- ============================================================
-- Jalankan sebelum setiap batch test agar data dalam keadaan bersih
-- dan deterministik. Bisa dijalankan berulang kali.
--
-- Usage:
--   psql "postgres://postgres:@localhost:5432/praktikum_backend?sslmode=disable" -f database/test_setup.sql
-- atau via run-tests.ps1 / run-tests.sh
-- ============================================================

-- 1. Hapus semua data & reset sequence SERIAL (id mulai dari 1 lagi)
TRUNCATE TABLE students RESTART IDENTITY;

-- 2. Seed data deterministik (id 1..4)
--    id=1  → dipakai sebagai studentId awal (environment studentId=1)
--    id=3  → target tes PUT 409 (NIM konflik)
--    id=4  → NIM 434241047 yang dipakai oleh tes 409 POST & PUT
INSERT INTO students (nim, name, grade, is_active) VALUES
    ('434241001', 'Ahmad Setiawan', 3.7, true),   -- id=1 (studentId awal)
    ('434241002', 'Budi Santoso', 3.3, true),      -- id=2
    ('434241003', 'Citra Lestari', 3.9, true),    -- id=3 (target PUT 409)
    ('434241047', 'Mahasiswa Ketiga', 3.5, true); -- id=4 (NIM konflik)

-- 3. Verifikasi hasil seed (opsional, untuk debugging manual)
-- SELECT id, nim, name, grade, is_active FROM students ORDER BY id;