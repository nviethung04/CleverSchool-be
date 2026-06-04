-- Migration: Recreate comments tables with id as primary key (ROLLBACK)
-- Date: 2025-01-08

-- Drop the new tables
DROP TABLE IF EXISTS homework_comments CASCADE;
DROP TABLE IF EXISTS exercise_comments CASCADE;
DROP TABLE IF EXISTS exam_comments CASCADE;

-- Recreate original tables (without id as primary key)
CREATE TABLE homework_comments (
    homework_id bigint NOT NULL,
    student_id bigint NOT NULL,
    teacher_id bigint NOT NULL,
    content text COLLATE pg_catalog."default",
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

CREATE TABLE exercise_comments (
    exercise_id bigint NOT NULL DEFAULT nextval('exercise_comments_exercise_id_seq'::regclass),
    student_id bigint NOT NULL,
    teacher_id bigint,
    content text COLLATE pg_catalog."default",
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

CREATE TABLE exam_comments (
    exam_id bigint NOT NULL DEFAULT nextval('exam_comments_exam_id_seq'::regclass),
    student_id bigint NOT NULL,
    teacher_id bigint,
    content text COLLATE pg_catalog."default",
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);
