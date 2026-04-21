-- Add provider column to meeting_notifications table
ALTER TABLE meeting_notifications 
ADD COLUMN provider VARCHAR(50) DEFAULT 'google' NOT NULL;

-- Create index on provider for faster queries
CREATE INDEX idx_meeting_notifications_provider ON meeting_notifications(provider);

-- Update existing records to have provider based on context
-- This will help identify which meetings are from which provider
UPDATE meeting_notifications 
SET provider = 'google' 
WHERE provider = 'google';
