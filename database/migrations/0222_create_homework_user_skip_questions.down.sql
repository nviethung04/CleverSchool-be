-- Drop indexes
DROP INDEX IF EXISTS idx_homework_user_skip_questions_did_it_again;
DROP INDEX IF EXISTS idx_homework_user_skip_questions_question_id;
DROP INDEX IF EXISTS idx_homework_user_skip_questions_user_id;
DROP INDEX IF EXISTS idx_homework_user_skip_questions_homework_id;

-- Drop foreign key constraints
ALTER TABLE public.homework_user_skip_questions DROP CONSTRAINT IF EXISTS fk_homework_user_skip_questions_question_id;
ALTER TABLE public.homework_user_skip_questions DROP CONSTRAINT IF EXISTS fk_homework_user_skip_questions_user_id;
ALTER TABLE public.homework_user_skip_questions DROP CONSTRAINT IF EXISTS fk_homework_user_skip_questions_homework_id;

-- Drop table
DROP TABLE IF EXISTS public.homework_user_skip_questions;

-- Drop sequence
DROP SEQUENCE IF EXISTS homework_user_skip_questions_id_seq;
