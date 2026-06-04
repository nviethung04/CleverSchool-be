DROP INDEX IF EXISTS idx_holidays_semester_id;

ALTER TABLE holidays DROP COLUMN IF EXISTS semester_id;
