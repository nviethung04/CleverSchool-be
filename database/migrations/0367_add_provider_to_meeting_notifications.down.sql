-- Rollback: Remove provider column from meeting_notifications table
ALTER TABLE meeting_notifications 
DROP COLUMN IF EXISTS provider;

-- Drop the index
DROP INDEX IF EXISTS idx_meeting_notifications_provider;
