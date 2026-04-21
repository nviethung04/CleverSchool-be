DROP TABLE IF EXISTS public.questions_homeworks_users;

CREATE TABLE IF NOT EXISTS public.homework_question_users
(
    id bigserial NOT NULL,
    homework_id bigint,
    user_id bigint,
    question_id bigint,
    answer_id bigint,
    true_answer_id bigint,
    is_correct boolean,
    score numeric(5,2),
    CONSTRAINT homework_question_users_pkey PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS public.exam_question_users
(
    id bigserial NOT NULL,
    exam_id bigint,
    user_id bigint,
    question_id bigint,
    answer_id bigint,
    true_answer_id bigint,
    is_correct boolean,
    score numeric(5,2),
    CONSTRAINT exam_question_users_pkey PRIMARY KEY (id)
    );
