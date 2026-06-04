-- 1. Thêm cột program_id nếu chưa có
ALTER TABLE courses ADD COLUMN IF NOT EXISTS program_id INT;
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS program_id INT;
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS program_id INT;
ALTER TABLE lesson_plans ADD COLUMN IF NOT EXISTS program_id INT;
ALTER TABLE lesson_plan_parts ADD COLUMN IF NOT EXISTS program_id INT;
ALTER TABLE homeworks ADD COLUMN IF NOT EXISTS program_id INT;
ALTER TABLE exams ADD COLUMN IF NOT EXISTS program_id INT;
ALTER TABLE cloned_questions ADD COLUMN IF NOT EXISTS program_id INT;
