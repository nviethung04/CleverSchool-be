ALTER TABLE google_meetings
    DROP COLUMN IF EXISTS short_code,
    DROP COLUMN IF EXISTS recording_url,
    DROP COLUMN IF EXISTS recording_status,
    DROP COLUMN IF EXISTS recording_duration_minutes;

DROP INDEX IF EXISTS idx_google_meetings_short_code;
