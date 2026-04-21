ALTER TABLE notification_logs
    DROP COLUMN user_id,
    DROP COLUMN device_token,
    DROP COLUMN error_message,
    DROP COLUMN fcm_message_id,
    DROP COLUMN read_at,
    DROP COLUMN delivered_at;

ALTER TABLE notification_logs
    ADD COLUMN data JSONB DEFAULT '{}'::jsonb;
