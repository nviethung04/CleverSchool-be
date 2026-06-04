ALTER TABLE cloned_questions DROP COLUMN IF EXISTS program_id;
ALTER TABLE exams DROP COLUMN IF EXISTS program_id;
ALTER TABLE homeworks DROP COLUMN IF EXISTS program_id;
ALTER TABLE lesson_plan_parts DROP COLUMN IF EXISTS program_id;
ALTER TABLE lesson_plans DROP COLUMN IF EXISTS program_id;
ALTER TABLE lessons DROP COLUMN IF EXISTS program_id;
ALTER TABLE chapters DROP COLUMN IF EXISTS program_id;
ALTER TABLE courses DROP COLUMN IF EXISTS program_id;
