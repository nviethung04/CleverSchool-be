-- Create sequence for homework_user_skip_questions table
CREATE SEQUENCE IF NOT EXISTS homework_user_skip_questions_id_seq;

-- Create homework_user_skip_questions table
CREATE TABLE IF NOT EXISTS public.homework_user_skip_questions
(
    id bigint NOT NULL DEFAULT nextval('homework_user_skip_questions_id_seq'::regclass),
    homework_id bigint,
    user_id bigint,
    question_id bigint,
    did_it_again boolean,
    did_it_again_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    CONSTRAINT homework_user_skip_questions_pkey PRIMARY KEY (id)
);

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_homework_user_skip_questions_homework_id ON public.homework_user_skip_questions(homework_id);
CREATE INDEX IF NOT EXISTS idx_homework_user_skip_questions_user_id ON public.homework_user_skip_questions(user_id);
CREATE INDEX IF NOT EXISTS idx_homework_user_skip_questions_question_id ON public.homework_user_skip_questions(question_id);
CREATE INDEX IF NOT EXISTS idx_homework_user_skip_questions_did_it_again ON public.homework_user_skip_questions(did_it_again);
