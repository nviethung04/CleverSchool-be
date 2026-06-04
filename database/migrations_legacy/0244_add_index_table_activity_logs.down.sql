-- Xoá các index (nếu tồn tại)
DROP INDEX IF EXISTS idx_activity_logs_device_created_at;
DROP INDEX IF EXISTS idx_activity_logs_device;
DROP INDEX IF EXISTS idx_activity_logs_user_created_at;
DROP INDEX IF EXISTS idx_activity_logs_user_id;
DROP INDEX IF EXISTS idx_activity_logs_created_at;

-- Xoá cột device
ALTER TABLE activity_logs
DROP COLUMN IF EXISTS device;
