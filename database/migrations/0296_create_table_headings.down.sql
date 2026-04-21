DROP TABLE IF EXISTS headings;

DROP INDEX IF EXISTS idx_lessons_heading_id;
ALTER TABLE lessons
    DROP COLUMN IF EXISTS heading_id;

ALTER TABLE chapters
    DROP COLUMN IF EXISTS time;

ALTER TABLE chapters
    ADD COLUMN time VARCHAR(255);
