-- 010: permission baru slice S-004 (ulasan kelas + detail user).
--
-- user:read        -> GET /users/:id (katalog helper.PermissionActions.UserRead)
-- review:moderate  -> moderasi ulasan (modul memakai di test; endpoint penuh S-007)
--
-- Idempotent (IF NOT EXISTS / ON CONFLICT DO NOTHING) supaya aman dijalankan ulang.
-- Diterapkan manual: psql -U postgres -d <db> -f migrations/010_review_permissions.sql
INSERT INTO permissions (name, description) VALUES
    ('user:read',       'Melihat detail satu user'),
    ('review:moderate', 'Memoderasi ulasan kelas')
ON CONFLICT (name) DO NOTHING;

-- Admin mendapat kedua permission baru.
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'user:read'),
    ('admin', 'review:moderate')
ON CONFLICT DO NOTHING;
