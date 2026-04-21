CREATE TABLE IF NOT EXISTS public.level_standards
(
    id bigserial NOT NULL,
    name bigint,
    create_by bigint,
    create_at timestamp without time zone,
    update_by bigint,
    update_at timestamp without time zone,
    deleted_by bigint,
    deleted_at timestamp without time zone,
    CONSTRAINT level_rules_pkey PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS public.level_test_question_users
(
    id bigserial NOT NULL,
    level_test_id bigint,
    user_id bigint,
    question_id bigint,
    answer_id bigint,
    true_answer_id bigint,
    is_correct boolean,
    score numeric(5,2),
    created_at timestamp without time zone,
    CONSTRAINT level_test_question_users_pkey PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS public.level_test_questions
(
    id bigserial NOT NULL,
    level_test_id bigint,
    question_id bigint,
    is_source_question boolean,
    score numeric(5,2),
    CONSTRAINT level_test_questions_pkey PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS public.level_tests
(
    id bigserial NOT NULL,
    name character varying(50) COLLATE pg_catalog."default",
    course_id bigint,
    level_standard_id bigint,
    status smallint,
    time_limit bigint,
    max_score numeric(5,2),
    description text COLLATE pg_catalog."default",
    cover_image text COLLATE pg_catalog."default",
    deadline timestamp without time zone,
    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint,
    CONSTRAINT level_tests_pkey PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS public.level_users
(
    id bigserial NOT NULL,
    user_id bigint,
    level_test_id bigint,
    score numeric(5,2),
    level_id bigint,
    created_at timestamp without time zone,
    CONSTRAINT level_users_pkey PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS public.levels
(
    id bigserial NOT NULL,
    level_standard_id bigint,
    name bigint,
    min_point numeric(5,2),
    max_point numeric(5,2),
    created_by bigint,
    created_at timestamp without time zone,
    updated_by bigint,
    updated_at timestamp without time zone,
    deleted_by bigint,
    deleted_at timestamp without time zone,
    CONSTRAINT levels_pkey PRIMARY KEY (id)
    )