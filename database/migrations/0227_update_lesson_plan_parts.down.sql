DROP INDEX IF EXISTS idx_lesson_plan_parts_course_id;
ALTER TABLE lesson_plan_parts DROP COLUMN IF EXISTS course_id;
