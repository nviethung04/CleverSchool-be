ALTER TABLE courses DROP COLUMN IF EXISTS clone_info;

ALTER TABLE chapters DROP COLUMN IF EXISTS clone_info;

ALTER TABLE lessons DROP COLUMN IF EXISTS clone_info;

ALTER TABLE lesson_plans DROP COLUMN IF EXISTS clone_info;

ALTER TABLE lesson_plan_parts DROP COLUMN IF EXISTS clone_info;

ALTER TABLE exams DROP COLUMN IF EXISTS clone_info;

ALTER TABLE homeworks DROP COLUMN IF EXISTS clone_info;

ALTER TABLE cloned_questions DROP COLUMN IF EXISTS clone_info;
