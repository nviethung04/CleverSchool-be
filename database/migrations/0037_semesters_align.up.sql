-- Align semesters table with models/prot expectations.
-- Some environments were created from older scheduling migration without these columns.

ALTER TABLE semesters
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS begin_date TIMESTAMP,
    ADD COLUMN IF NOT EXISTS status BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS sort_order INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS previous_semester_id BIGINT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS start_week_id BIGINT,
    ADD COLUMN IF NOT EXISTS end_week_id BIGINT;

-- Backfill begin_date when missing
UPDATE semesters
SET begin_date = COALESCE(begin_date, start_date)
WHERE begin_date IS NULL;

