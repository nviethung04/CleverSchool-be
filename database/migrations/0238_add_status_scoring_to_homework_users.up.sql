-- Add status_scoring column to homework_users table
ALTER TABLE homework_users 
ADD COLUMN IF NOT EXISTS status_scoring SMALLINT DEFAULT 0;

-- Add comment to explain the status values
COMMENT ON COLUMN homework_users.status_scoring IS '0: không cần chấm, 1: chưa chấm xong, 2: đã chấm xong';

-- Add index for better performance when querying by status_scoring
CREATE INDEX IF NOT EXISTS idx_homework_users_status_scoring 
ON homework_users(status_scoring);
