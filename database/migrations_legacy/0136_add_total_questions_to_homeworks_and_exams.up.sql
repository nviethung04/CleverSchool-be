ALTER TABLE homeworks ADD COLUMN IF NOT EXISTS total_questions bigint;
ALTER TABLE exams ADD COLUMN IF NOT EXISTS total_questions bigint; 