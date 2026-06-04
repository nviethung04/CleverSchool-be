CREATE TABLE IF NOT EXISTS public.exercise_question_user_labelings
(
    id bigserial NOT NULL,
    exercise_id bigint,
    question_id bigint,
    user_id bigint,
    blank_id bigint,
    answer_id bigint,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT exercise_question_user_labelings_pkey PRIMARY KEY (id)
);
