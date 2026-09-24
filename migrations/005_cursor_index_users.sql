-- 005: index untuk cursor pagination pada tabel users.
-- Arah index DESC harus cocok dengan klausa ORDER BY DESC di
-- repository agar planner memilih Index Scan, bukan Seq Scan
-- (Modul 7 bug #8).
CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx
ON users (created_at DESC, id DESC);
