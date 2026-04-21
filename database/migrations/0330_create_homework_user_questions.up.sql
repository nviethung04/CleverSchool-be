-- Create homework_user_questions table
CREATE TABLE IF NOT EXISTS public.homework_user_questions
(
    id bigserial,
    homework_id bigint,
    user_id bigint,
    question_id bigint,
    ratio_score numeric(6,3),
    is_all_correct boolean,
    star integer,
    number_options integer,
    created_at timestamp without time zone,
    lesson_id bigint,
    number_time_sent integer,
    CONSTRAINT homework_user_questions_pkey PRIMARY KEY (id)
);

-- Indexes for common lookup columns
CREATE INDEX IF NOT EXISTS idx_homework_user_questions_homework_id ON public.homework_user_questions(homework_id);
CREATE INDEX IF NOT EXISTS idx_homework_user_questions_user_id ON public.homework_user_questions(user_id);
CREATE INDEX IF NOT EXISTS idx_homework_user_questions_question_id ON public.homework_user_questions(question_id);
CREATE INDEX IF NOT EXISTS idx_homework_user_questions_lesson_id ON public.homework_user_questions(lesson_id);

