-- Migration: Recreate comments tables with id as primary key
-- Date: 2025-01-08

-- Drop existing comments tables if they exist
DROP TABLE IF EXISTS homework_comments CASCADE;
DROP TABLE IF EXISTS exercise_comments CASCADE;
DROP TABLE IF EXISTS exam_comments CASCADE;

-- Create homework_comments table with id as primary key
CREATE TABLE homework_comments (
    id bigserial PRIMARY KEY,
    homework_id bigint NOT NULL,
    student_id bigint NOT NULL,
    teacher_id bigint NOT NULL,
    content text COLLATE pg_catalog."default",
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint,
    lesson_id bigint
);

-- Create exercise_comments table with id as primary key
CREATE TABLE exercise_comments (
    id bigserial PRIMARY KEY,
    exercise_id bigint NOT NULL,
    student_id bigint NOT NULL,
    teacher_id bigint,
    content text COLLATE pg_catalog."default",
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint,
    lesson_id bigint
);

-- Create exam_comments table with id as primary key
CREATE TABLE exam_comments (
    id bigserial PRIMARY KEY,
    exam_id bigint NOT NULL,
    student_id bigint NOT NULL,
    teacher_id bigint,
    content text COLLATE pg_catalog."default",
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint,
    lesson_id bigint
);

-- Add indexes for better performance
CREATE INDEX idx_homework_comments_homework_student ON homework_comments(homework_id, student_id);
CREATE INDEX idx_homework_comments_deleted_at ON homework_comments(deleted_at);

CREATE INDEX idx_exercise_comments_exercise_student ON exercise_comments(exercise_id, student_id);
CREATE INDEX idx_exercise_comments_deleted_at ON exercise_comments(deleted_at);

CREATE INDEX idx_exam_comments_exam_student ON exam_comments(exam_id, student_id);
CREATE INDEX idx_exam_comments_deleted_at ON exam_comments(deleted_at);
