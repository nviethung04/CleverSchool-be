ALTER TABLE exams DROP COLUMN IF EXISTS is_random_question;

ALTER TABLE homeworks DROP COLUMN IF EXISTS is_random_question;

ALTER TABLE exercises DROP COLUMN IF EXISTS is_random_question;
