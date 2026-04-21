-- Create user_devices table for storing FCM tokens
CREATE TABLE IF NOT EXISTS user_devices (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    device_token VARCHAR(500) NOT NULL UNIQUE,
    device_type VARCHAR(20) NOT NULL,
    device_name VARCHAR(100),
    platform VARCHAR(20),
    app_version VARCHAR(20),
    os_version VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices(user_id);
CREATE INDEX IF NOT EXISTS idx_user_devices_token ON user_devices(device_token);
CREATE INDEX IF NOT EXISTS idx_user_devices_active ON user_devices(is_active);
CREATE INDEX IF NOT EXISTS idx_user_devices_type ON user_devices(device_type);
CREATE INDEX IF NOT EXISTS idx_user_devices_deleted_at ON user_devices(deleted_at);

-- Add comments
COMMENT ON TABLE user_devices IS 'Store user device tokens for push notifications (FCM)';
COMMENT ON COLUMN user_devices.device_type IS 'Device type: ios, android, web';
COMMENT ON COLUMN user_devices.device_token IS 'FCM (Firebase Cloud Messaging) token';
COMMENT ON COLUMN user_devices.is_active IS 'Whether the device is still active (not logged out or uninstalled)';
COMMENT ON COLUMN user_devices.deleted_at IS 'Soft delete timestamp';
