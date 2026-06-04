ALTER TABLE activity_logs
ADD COLUMN device VARCHAR(20);

UPDATE activity_logs
SET device = CASE
    WHEN agent ILIKE '%iPhone%' THEN 'Mobile'
    WHEN agent ILIKE '%iPad%' THEN 'Tablet'
    WHEN agent ILIKE '%Android%' AND agent ILIKE '%Mobile%' THEN 'Mobile'
    WHEN agent ILIKE '%Android%' THEN 'Tablet'
    WHEN agent ILIKE '%Windows NT%' THEN 'Desktop'
    WHEN agent ILIKE '%Macintosh%' THEN 'Desktop'
    ELSE 'Tablet'
END;

CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_created_at ON activity_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_device ON activity_logs(device);
CREATE INDEX IF NOT EXISTS idx_activity_logs_device_created_at ON activity_logs(device, created_at DESC);
