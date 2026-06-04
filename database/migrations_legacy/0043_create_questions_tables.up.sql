-- XÓA BẢNG CŨ (nếu tồn tại)
DROP TABLE IF EXISTS public.questions_homeworks;
DROP TABLE IF EXISTS public.questions_exams;
DROP TABLE IF EXISTS public.lesson_plan_parts_questions;

-- TẠO BẢNG homeworks_questions
CREATE TABLE IF NOT EXISTS public.homeworks_questions
(
    id bigserial NOT NULL,
    homework_id bigint,
    question_id bigint,
    CONSTRAINT homeworks_questions_pkey PRIMARY KEY (id)
    );

-- TẠO BẢNG exams_questions
CREATE TABLE IF NOT EXISTS public.exams_questions
(
    id bigserial NOT NULL,
    exam_id bigint,
    question_id bigint,
    CONSTRAINT exams_questions_pkey PRIMARY KEY (id)
    );

-- TẠO BẢNG lesson_plan_parts_questions
CREATE TABLE IF NOT EXISTS public.lesson_plan_parts_questions
(
    id bigserial NOT NULL,
    lesson_plan_part_id bigint,
    question_id bigint,
    CONSTRAINT lesson_plan_parts_questions_pkey PRIMARY KEY (id)
    );
