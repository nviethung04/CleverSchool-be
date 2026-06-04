-- +migrate Down

-- Drop indexes created in 0185
DROP INDEX IF EXISTS public.uq_exercise_ref_lessons;
DROP INDEX IF EXISTS public.idx_exercise_ref_lessons_lesson_id;
DROP INDEX IF EXISTS public.idx_exercise_ref_lessons_homework_id;
DROP INDEX IF EXISTS public.exercises_questions_pkey;
DROP INDEX IF EXISTS public.exercise_question_user_files_pkey;

-- Drop tables created in 0185 (reverse order)
DROP TABLE IF EXISTS public.excercise_comments;
DROP TABLE IF EXISTS public.exercise_question_user_fill_in_blanks;
DROP TABLE IF EXISTS public.exercise_question_user_groups;
DROP TABLE IF EXISTS public.exercise_question_user_manual_scoring;
DROP TABLE IF EXISTS public.exercise_question_user_matchings;
DROP TABLE IF EXISTS public.exercise_question_user_positions;
DROP TABLE IF EXISTS public.exercise_question_users;
DROP TABLE IF EXISTS public.exercise_questions;
DROP TABLE IF EXISTS public.exercise_ref_lessons;
DROP TABLE IF EXISTS public.exercise_users;
DROP TABLE IF EXISTS public.exercises;


