DROP INDEX IF EXISTS idx_microsoft_meetings_short_code;
DROP INDEX IF EXISTS idx_microsoft_meetings_class_id;

ALTER TABLE microsoft_meetings
    DROP COLUMN IF EXISTS short_code,
    DROP COLUMN IF EXISTS recording_url,
    DROP COLUMN IF EXISTS recording_status,
    DROP COLUMN IF EXISTS class_id;
