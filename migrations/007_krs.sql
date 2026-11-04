BEGIN;

ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;

CREATE TABLE IF NOT EXISTS courses (
    id SERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    credits SMALLINT NOT NULL CHECK (credits BETWEEN 1 AND 6),
    quota INTEGER NOT NULL CHECK (quota > 0),
    enrolled_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (enrolled_count BETWEEN 0 AND quota)
);

CREATE TABLE IF NOT EXISTS enrollments (
    id BIGSERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (course_id, user_id)
);

CREATE INDEX IF NOT EXISTS enrollments_user_idx ON enrollments (user_id);

INSERT INTO courses (code, name, credits, quota)
SELECT seed.code, seed.name, 3, 40
FROM (VALUES
    (1, 'IF101', 'Pemrograman Backend Lanjut'),
    (2, 'IF102', 'Basis Data Lanjut'),
    (3, 'IF301', 'Kecerdasan Buatan'),
    (4, 'IF302', 'Keamanan Informasi'),
    (5, 'IF303', 'Komputasi Awan'),
    (6, 'IF304', 'Data Mining'),
    (7, 'IF305', 'Pemrograman Mobile'),
    (8, 'IF306', 'Interaksi Manusia Komputer'),
    (9, 'IF307', 'Sistem Terdistribusi')
) AS seed(seq, code, name)
WHERE NOT EXISTS (SELECT 1 FROM courses c WHERE c.code = seed.code)
ORDER BY seed.seq
ON CONFLICT (code) DO NOTHING;

INSERT INTO permissions (name, description) VALUES
    ('course:enroll', 'Mengelola KRS sendiri'),
    ('course:manage', 'Mengubah kuota kelas')
ON CONFLICT DO NOTHING;

INSERT INTO roles (name, description)
VALUES ('user', 'Akun registrasi mandiri')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('user', 'course:enroll'),
    ('admin', 'course:manage')
ON CONFLICT DO NOTHING;

COMMIT;
