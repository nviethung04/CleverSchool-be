DROP TABLE IF EXISTS public.exam_question_users;

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
    created_at timestamp without time zone,
    CONSTRAINT exam_question_users_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.exam_question_user_positions;

CREATE TABLE IF NOT EXISTS public.exam_question_user_positions
(
    id bigserial NOT NULL,
    exam_id bigint,
    user_id bigint,
    question_id bigint,
    answer_id bigint,
    sort_position integer,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT exam_question_user_positions_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.exam_question_user_matchings;

CREATE TABLE IF NOT EXISTS public.exam_question_user_matchings
(
    id bigserial NOT NULL,
    exam_id bigint,
    user_id bigint,
    question_id bigint,
    first_item_id bigint,
    second_item_id bigint,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT exam_question_user_matchings_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.exam_question_user_labelings;

CREATE TABLE IF NOT EXISTS public.exam_question_user_labelings
(
    id bigserial NOT NULL,
    exam_id bigint,
    question_id bigint,
    user_id bigint,
    blank_id bigint,
    answer_id bigint,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT exam_question_user_labelings_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.exam_question_user_groups;

CREATE TABLE IF NOT EXISTS public.exam_question_user_groups
(
    id bigserial NOT NULL,
    exam_id bigint,
    user_id bigint,
    question_id bigint,
    answer_id bigint,
    group_id bigint,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT exam_question_user_groups_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.exam_question_user_fill_in_blanks;

CREATE TABLE IF NOT EXISTS public.exam_question_user_fill_in_blanks
(
    id bigserial NOT NULL,
    exam_id bigint,
    user_id bigint,
    question_id bigint,
    answer text COLLATE pg_catalog."default",
    sort_position integer,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT exam_question_user_fill_in_blanks_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.exam_users;

CREATE TABLE IF NOT EXISTS public.exam_users
(
    id bigserial NOT NULL,
    exam_id bigint,
    user_id bigint,
    score numeric(5,2),
    ratio numeric(5,2),
    "time" bigint,
    is_score boolean,
    CONSTRAINT exam_users_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.homework_question_users;

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
    created_at timestamp without time zone,
    CONSTRAINT homework_question_users_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.homework_question_user_positions;

CREATE TABLE IF NOT EXISTS public.homework_question_user_positions
(
    id bigserial NOT NULL,
    homework_id bigint,
    user_id bigint,
    question_id bigint,
    answer_id bigint,
    sort_position integer,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT homework_question_user_positions_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.homework_question_user_matchings;

CREATE TABLE IF NOT EXISTS public.homework_question_user_matchings
(
    id bigserial NOT NULL,
    homework_id bigint,
    user_id bigint,
    question_id bigint,
    first_item_id bigint,
    second_item_id bigint,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT homework_question_user_matchings_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.homework_question_user_labelings;

CREATE TABLE IF NOT EXISTS public.homework_question_user_labelings
(
    id bigserial NOT NULL,
    homework_id bigint,
    question_id bigint,
    user_id bigint,
    blank_id bigint,
    answer_id bigint,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT homework_question_user_labelings_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.homework_question_user_groups;

CREATE TABLE IF NOT EXISTS public.homework_question_user_groups
(
    id bigserial NOT NULL,
    homework_id bigint,
    user_id bigint,
    question_id bigint,
    answer_id bigint,
    group_id bigint,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT homework_question_user_groups_pkey PRIMARY KEY (id)
    );

DROP TABLE IF EXISTS public.homework_question_user_fill_in_blanks;

CREATE TABLE IF NOT EXISTS public.homework_question_user_fill_in_blanks
(
    id bigserial NOT NULL,
    homework_id bigint,
    user_id bigint,
    question_id bigint,
    answer text COLLATE pg_catalog."default",
    sort_position integer,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT homework_question_user_fill_in_blanks_pkey PRIMARY KEY (id)
    );