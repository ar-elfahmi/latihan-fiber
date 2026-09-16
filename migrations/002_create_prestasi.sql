CREATE TABLE IF NOT EXISTS prestasi (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL,
    keterangan VARCHAR(100) NOT NULL,
    juara INTEGER NOT NULL CHECK (juara >= 0 AND juara <= 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
    CONSTRAINT fk_prestasi_student
        FOREIGN KEY (student_id)
        REFERENCES students (id)
        ON DELETE CASCADE
);