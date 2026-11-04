-- Reset C.1 dan C.3 pada database uji Modul 11.
\set ON_ERROR_STOP on
BEGIN;

DO $$
BEGIN
    IF current_database() <> 'praktikum_modul11_uji' THEN
        RAISE EXCEPTION 'Reset Modul 11 hanya boleh dijalankan pada praktikum_modul11_uji';
    END IF;
END;
$$;

DELETE FROM notifications
WHERE user_id IN (SELECT id FROM users WHERE username IN ('mhsuji', 'salahalamat'));

DELETE FROM outbox
WHERE (payload->>'user_id')::INT IN (
    SELECT id FROM users WHERE username IN ('mhsuji', 'salahalamat')
);

DELETE FROM enrollments
WHERE user_id IN (SELECT id FROM users WHERE username IN ('mhsuji', 'salahalamat'));

UPDATE courses c
SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id);

COMMIT;
