-- +migrate Up

-- Drop wrong table if exists
DROP TABLE IF EXISTS public.excercise_comments CASCADE;

-- Create correct table
CREATE TABLE IF NOT EXISTS public.exercise_comments
(
    exercises_id BIGINT NOT NULL,
    student_id   BIGINT NOT NULL,
    teacher_id   BIGINT NOT NULL,
    content      TEXT,
    created_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_by   BIGINT,
    deleted_at   TIMESTAMP WITHOUT TIME ZONE,
    deleted_by   BIGINT
);

CREATE INDEX IF NOT EXISTS idx_exercise_comments_exercises_id ON public.exercise_comments (exercises_id);
CREATE INDEX IF NOT EXISTS idx_exercise_comments_student_id ON public.exercise_comments (student_id);
CREATE INDEX IF NOT EXISTS idx_exercise_comments_teacher_id ON public.exercise_comments (teacher_id);


