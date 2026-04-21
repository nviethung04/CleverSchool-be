-- Drop indexes
DROP INDEX IF EXISTS idx_homework_user_questions_question_id;
DROP INDEX IF EXISTS idx_homework_user_questions_user_id;
DROP INDEX IF EXISTS idx_homework_user_questions_homework_id;

-- Drop table
DROP TABLE IF EXISTS public.homework_user_questions;

-- Drop sequence
DROP SEQUENCE IF EXISTS homework_user_questions_id_seq;

