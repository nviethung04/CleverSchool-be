DROP INDEX IF EXISTS idx_classes_grade_id;

ALTER TABLE
    classes DROP COLUMN IF EXISTS grade_id;
