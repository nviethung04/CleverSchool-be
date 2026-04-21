ALTER TABLE microsoft_meetings
ADD COLUMN IF NOT EXISTS recording_share_url TEXT;
