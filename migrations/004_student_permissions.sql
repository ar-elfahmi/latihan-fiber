-- 004: skema RBAC + permission mahasiswa + kolom owner_id.
--
-- Isi file ini (urutan penting):
--   1. Empat tabel RBAC: roles, permissions, role_permissions, user_roles.
--   2. Seed permission student:* beserta pasangannya ke role (tabel di tugas).
--   3. Kolom owner_id pada students: dasar pemeriksaan kepemilikan.
--   4. Akun demo admin/staff/user + contoh data mahasiswa milik masing-masing.
--
-- Semua statement idempotent (IF NOT EXISTS / ON CONFLICT DO NOTHING)
-- sehingga aman dijalankan ulang. Migration yang gagal di tengah jalan
-- lebih merepotkan daripada migration yang ditulis hati-hati sejak awal.

-- =============================================================
-- 1. Tabel RBAC (many-to-many: role<->permission, user<->role)
-- =============================================================
CREATE TABLE IF NOT EXISTS roles (
    name        TEXT PRIMARY KEY,
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS permissions (
    name        TEXT PRIMARY KEY,
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       TEXT NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name TEXT NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id   INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_name TEXT NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_name)
);

-- =============================================================
-- 2. Seed role, permission, dan pemasangan role -> permission
-- =============================================================
INSERT INTO roles (name, description) VALUES
    ('admin', 'Akses penuh: kelola mahasiswa & user lain.'),
    ('staff', 'Akses operasional: lihat & input mahasiswa.'),
    ('user',  'Akses terbatas: hanya data miliknya sendiri.')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description) VALUES
    -- manajemen user (dipasang di praktikum, disertakan agar lengkap)
    ('user:list',           'Melihat daftar seluruh user'),
    ('user:read:any',       'Melihat data user mana pun'),
    ('user:update:any',     'Mengubah data user mana pun'),
    ('user:delete',         'Menghapus user'),
    ('role:assign',         'Mengubah role milik user lain'),
    -- student (inti Tugas 6 C.1)
    ('student:list',        'Melihat daftar mahasiswa'),
    ('student:read:any',    'Melihat mahasiswa mana pun'),
    ('student:read:self',   'Melihat mahasiswa miliknya sendiri'),
    ('student:create',      'Membuat mahasiswa baru'),
    ('student:update:any',  'Mengubah mahasiswa mana pun'),
    ('student:update:self', 'Mengubah mahasiswa miliknya sendiri'),
    ('student:delete',      'Menghapus mahasiswa'),
    -- prestasi
    ('prestasi:list',       'Melihat daftar prestasi'),
    ('prestasi:read:any',   'Melihat prestasi mahasiswa mana pun'),
    ('prestasi:read:self',  'Melihat prestasi miliknya sendiri'),
    ('prestasi:create',     'Membuat prestasi baru')
ON CONFLICT (name) DO NOTHING;

-- Pemasangan sesuai tabel kebijakan di tugas:
--   student:list / student:read:any / student:create -> admin, staff
--   student:update:any / student:delete              -> admin saja
-- Role user diberi pasangan :self supaya pemilik data bisa lolos
-- middleware RequirePermission lalu diperiksa kepemilikannya di service.
INSERT INTO role_permissions (role_name, permission_name) VALUES
    -- admin: hak penuh
    ('admin', 'user:list'),
    ('admin', 'user:read:any'),
    ('admin', 'user:update:any'),
    ('admin', 'user:delete'),
    ('admin', 'role:assign'),
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('admin', 'prestasi:list'),
    ('admin', 'prestasi:read:any'),
    ('admin', 'prestasi:create'),
    -- staff: operasional (tanpa update:any, tanpa delete)
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create'),
    ('staff', 'prestasi:list'),
    ('staff', 'prestasi:read:any'),
    ('staff', 'prestasi:create'),
    -- user: hanya miliknya sendiri
    ('user', 'student:read:self'),
    ('user', 'student:update:self'),
    ('user', 'prestasi:read:self')
ON CONFLICT DO NOTHING;

-- =============================================================
-- 3. Kolom owner_id pada students
-- =============================================================
-- Baris lama yang belum punya pemilik dibiarkan NULL, bukan dipaksa
-- NOT NULL. Alasannya: NULL berarti "belum ada pemilik", sehingga baris
-- tersebut hanya bisa diakses pemegang permission :any. Kalau dipaksa
-- NOT NULL, migration akan gagal di tabel yang sudah berisi data
-- (tidak ada nilai default yang jujur untuk kepemilikan).
-- Kalau versi lama kolom owner_user_id sudah terlanjur ada, ganti namanya
-- dulu supaya datanya tidak hilang, baru pastikan kolom owner_id ada.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'students' AND column_name = 'owner_user_id')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'students' AND column_name = 'owner_id') THEN
        ALTER TABLE students RENAME COLUMN owner_user_id TO owner_id;
    END IF;
END $$;

ALTER TABLE students
    ADD COLUMN IF NOT EXISTS owner_id INT NULL REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);

-- =============================================================
-- 4. Akun demo + contoh data untuk pengujian matriks hak akses
-- =============================================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- admin/admin123 sudah dibuat di 003; tambahkan staff dan user.
INSERT INTO users (username, password_hash)
VALUES
    ('staff', crypt('staff123', gen_salt('bf', 10))),
    ('user',  crypt('user123',  gen_salt('bf', 10)))
ON CONFLICT (username) DO NOTHING;

INSERT INTO user_roles (user_id, role_name)
SELECT u.id, x.role
FROM (VALUES ('admin', 'admin'), ('staff', 'staff'), ('user', 'user')) AS x(username, role)
JOIN users u ON u.username = x.username
ON CONFLICT DO NOTHING;

-- Dua mahasiswa contoh: satu milik 'user' (untuk uji "pemilik"),
-- satu milik 'admin' (untuk uji "bukan pemilik"). NIM unik -> aman ulang.
INSERT INTO students (nim, name, grade, is_active, owner_id)
SELECT x.nim, x.name, x.grade, x.is_active, u.id
FROM (VALUES
    ('2101010101', 'Budi Santoso', 3.2::double precision, true, 'user'),
    ('2101010102', 'Sari Dewi',    3.5::double precision, true, 'admin')
) AS x(nim, name, grade, is_active, owner)
JOIN users u ON u.username = x.owner
ON CONFLICT (nim) DO NOTHING;
