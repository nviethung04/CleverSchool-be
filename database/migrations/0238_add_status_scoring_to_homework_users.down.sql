-- Remove index
DROP INDEX IF EXISTS idx_homework_users_status_scoring;

-- Remove status_scoring column from homework_users table
ALTER TABLE homework_users 
DROP COLUMN IF EXISTS status_scoring;
