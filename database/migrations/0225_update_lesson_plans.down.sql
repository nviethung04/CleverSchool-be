DROP INDEX IF EXISTS idx_lesson_plans_lesson_id;
ALTER TABLE lesson_plans DROP COLUMN IF EXISTS lesson_id;
