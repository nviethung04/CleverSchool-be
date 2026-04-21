-- Create notification_logs table for tracking push notification delivery
CREATE TABLE IF NOT EXISTS notification_logs (
    id SERIAL PRIMARY KEY,
    notice_id INT NOT NULL,
    user_id INT NOT NULL,
    device_token VARCHAR(500),
    status VARCHAR(20) DEFAULT 'pending',
    error_message TEXT,
    fcm_message_id VARCHAR(255),
    sent_at TIMESTAMP DEFAULT NOW(),
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_notification_logs_notice_id ON notification_logs(notice_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_user_id ON notification_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_status ON notification_logs(status);
CREATE INDEX IF NOT EXISTS idx_notification_logs_sent_at ON notification_logs(sent_at);

-- Add comments
COMMENT ON TABLE notification_logs IS 'Log all push notification attempts and delivery status';
COMMENT ON COLUMN notification_logs.status IS 'Status: pending, sent, delivered, failed, read';
COMMENT ON COLUMN notification_logs.fcm_message_id IS 'FCM message ID returned from Firebase';
