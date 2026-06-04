-- +migrate Up

-- exercises
CREATE TABLE IF NOT EXISTS public.exercises
(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50),
    lesson_id BIGINT,
    status SMALLINT,
    time_limit BIGINT,
    max_score REAL,
    description TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT,
    deadline TIMESTAMP WITHOUT TIME ZONE,
    is_assigned BOOLEAN,
    cover_image_info JSONB,
    total_questions BIGINT,
    assigned_at TIMESTAMP WITHOUT TIME ZONE,
    assigned_by BIGINT,
    clone_info JSONB,
    object_title VARCHAR(255),
    program_id INTEGER
);

-- exercise_users
CREATE TABLE IF NOT EXISTS public.exercise_users
(
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT,
    user_id BIGINT,
    score NUMERIC(5,2),
    ratio NUMERIC(5,2),
    "time" BIGINT,
    has_manual_scoring BOOLEAN,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    updated_at TIMESTAMP WITHOUT TIME ZONE
);

-- exercise_ref_lessons
CREATE TABLE IF NOT EXISTS public.exercise_ref_lessons
(
    id BIGSERIAL PRIMARY KEY,
    lesson_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- indexes for exercise_ref_lessons
CREATE INDEX IF NOT EXISTS idx_exercise_ref_lessons_homework_id
    ON public.exercise_ref_lessons USING btree (exercise_id ASC NULLS LAST);

CREATE INDEX IF NOT EXISTS idx_exercise_ref_lessons_lesson_id
    ON public.exercise_ref_lessons USING btree (lesson_id ASC NULLS LAST);

CREATE UNIQUE INDEX IF NOT EXISTS uq_exercise_ref_lessons
    ON public.exercise_ref_lessons USING btree (lesson_id ASC NULLS LAST, exercise_id ASC NULLS LAST);

-- exercise_questions
CREATE TABLE IF NOT EXISTS public.exercise_questions
(
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT,
    question_id BIGINT,
    score NUMERIC(5,2),
    is_source_question BOOLEAN DEFAULT FALSE
);

-- redundant unique index to match requested schema (will no-op if PK index exists)
CREATE UNIQUE INDEX IF NOT EXISTS exercises_questions_pkey
    ON public.exercise_questions USING btree (id ASC NULLS LAST);

-- exercise_question_users
CREATE TABLE IF NOT EXISTS public.exercise_question_users
(
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer_id BIGINT,
    true_answer_id BIGINT,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- exercise_question_user_positions
CREATE TABLE IF NOT EXISTS public.exercise_question_user_positions
(
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer_group_position BIGINT,
    sort_position INTEGER,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- exercise_question_user_matchings
CREATE TABLE IF NOT EXISTS public.exercise_question_user_matchings
(
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    first_item_id BIGINT,
    second_item_id BIGINT,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- exercise_question_user_manual_scoring
CREATE TABLE IF NOT EXISTS public.exercise_question_user_manual_scoring
(
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer TEXT,
    score NUMERIC(5,2),
    is_scored BOOLEAN,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    scoring_by BIGINT,
    scoring_at TIMESTAMP WITHOUT TIME ZONE,
    file_info JSONB
);

-- redundant unique index to match requested schema
CREATE UNIQUE INDEX IF NOT EXISTS exercise_question_user_files_pkey
    ON public.exercise_question_user_manual_scoring USING btree (id ASC NULLS LAST);

-- exercise_question_user_groups
CREATE TABLE IF NOT EXISTS public.exercise_question_user_groups
(
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer_id BIGINT,
    group_id BIGINT,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- exercise_question_user_fill_in_blanks
CREATE TABLE IF NOT EXISTS public.exercise_question_user_fill_in_blanks
(
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer TEXT,
    sort_position INTEGER,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- excercise_comments (note: name as requested)
CREATE TABLE IF NOT EXISTS public.excercise_comments
(
    excercises_id BIGSERIAL,
    student_id BIGINT NOT NULL,
    teacher_id BIGINT,
    content TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT
);


