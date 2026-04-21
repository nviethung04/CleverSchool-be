-- Create notices table for push notifications
CREATE TABLE IF NOT EXISTS notices (
    id SERIAL PRIMARY KEY,
    cover TEXT,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    content TEXT,
    status BOOLEAN NOT NULL DEFAULT FALSE,
    type VARCHAR(50) NOT NULL DEFAULT 'system',
    deleted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_notices_type ON notices(type);
CREATE INDEX IF NOT EXISTS idx_notices_status ON notices(status);

-- Add comments
COMMENT ON TABLE notices IS 'Push notifications for users';
COMMENT ON COLUMN notices.type IS 'Notification type: system, school, class, course, user';
COMMENT ON COLUMN notices.status IS 'true = published, false = draft';
