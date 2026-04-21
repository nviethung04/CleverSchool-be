-- Add missing columns used by application code
ALTER TABLE microsoft_meetings
    ADD COLUMN IF NOT EXISTS short_code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS recording_url TEXT,
    ADD COLUMN IF NOT EXISTS recording_status VARCHAR(20) DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS class_id INTEGER;

-- Indexes to improve lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_microsoft_meetings_short_code ON microsoft_meetings(short_code);
CREATE INDEX IF NOT EXISTS idx_microsoft_meetings_class_id ON microsoft_meetings(class_id);
