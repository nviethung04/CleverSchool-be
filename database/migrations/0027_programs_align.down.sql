DROP INDEX IF EXISTS idx_programs_subject_id;

ALTER TABLE programs
    DROP COLUMN IF EXISTS subject_id,
    DROP COLUMN IF EXISTS image_info,
    DROP COLUMN IF EXISTS target;
