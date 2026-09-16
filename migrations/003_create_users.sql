-- 003: tabel users untuk login JWT.
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Akun contoh untuk belajar: admin / admin123
-- Hash bcrypt dibuat langsung oleh PostgreSQL (pgcrypto).
CREATE EXTENSION IF NOT EXISTS pgcrypto;
INSERT INTO users (username, password_hash)
VALUES ('admin', crypt('admin123', gen_salt('bf', 10)))
ON CONFLICT (username) DO NOTHING;
