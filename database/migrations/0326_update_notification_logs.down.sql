ALTER TABLE notification_logs
    ADD COLUMN user_id BIGINT,
    ADD COLUMN device_token VARCHAR(500),
    ADD COLUMN error_message TEXT,
    ADD COLUMN fcm_message_id VARCHAR(255),
    ADD COLUMN read_at TIMESTAMP,
    ADD COLUMN delivered_at TIMESTAMP;

ALTER TABLE notification_logs
    DROP COLUMN data;
