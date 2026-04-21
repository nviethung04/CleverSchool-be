-- Add missing columns used by application code
ALTER TABLE google_meetings
    ADD COLUMN IF NOT EXISTS short_code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS recording_url TEXT,
    ADD COLUMN IF NOT EXISTS recording_status VARCHAR(20) DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS recording_duration_minutes INTEGER DEFAULT 0;

-- Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_google_meetings_short_code ON google_meetings(short_code);
