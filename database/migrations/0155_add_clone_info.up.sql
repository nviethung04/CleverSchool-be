ALTER TABLE courses ADD COLUMN clone_info JSONB;

ALTER TABLE chapters ADD COLUMN clone_info JSONB;

ALTER TABLE lessons ADD COLUMN clone_info JSONB;

ALTER TABLE lesson_plans ADD COLUMN clone_info JSONB;

ALTER TABLE lesson_plan_parts ADD COLUMN clone_info JSONB;

ALTER TABLE exams ADD COLUMN clone_info JSONB;

ALTER TABLE homeworks ADD COLUMN clone_info JSONB;

ALTER TABLE cloned_questions ADD COLUMN clone_info JSONB;
