BEGIN;

ALTER TABLE notifications
    DROP CONSTRAINT notifications_outbox_unique,
    ADD CONSTRAINT notifications_outbox_user_unique UNIQUE (outbox_id, user_id);

COMMIT;
