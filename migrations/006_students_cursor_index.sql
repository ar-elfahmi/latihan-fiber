-- 006: index untuk cursor pagination pada tabel students.
-- Pilih id saja (SERIAL PRIMARY KEY) sebagai column pengurut karena
-- nilainya selalu unik dan total-terurut tanpa tie-breaker. Arah DESC
-- cocok dengan ORDER BY di repository agar Index Scan dipakai.
CREATE INDEX IF NOT EXISTS students_id_desc_idx
ON students (id DESC);
