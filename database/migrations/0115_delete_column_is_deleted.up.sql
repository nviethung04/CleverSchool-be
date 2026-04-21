ALTER TABLE exams
DROP COLUMN IF EXISTS is_deleted;

ALTER TABLE homeworks
DROP COLUMN IF EXISTS is_deleted;

ALTER TABLE lesson_plan_parts
DROP COLUMN IF EXISTS is_deleted;

ALTER TABLE lesson_plans
DROP COLUMN IF EXISTS is_deleted;
