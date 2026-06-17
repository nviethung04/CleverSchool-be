-- Best-effort rollback (data in new columns may be lost).

ALTER TABLE semesters
    DROP COLUMN IF EXISTS end_week_id,
    DROP COLUMN IF EXISTS start_week_id,
    DROP COLUMN IF EXISTS previous_semester_id,
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS begin_date,
    DROP COLUMN IF EXISTS description;
