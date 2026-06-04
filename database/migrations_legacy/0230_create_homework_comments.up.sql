CREATE TABLE IF NOT EXISTS public.homework_comments (
    homework_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    teacher_id BIGINT NOT NULL,
    content TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT
);

CREATE INDEX IF NOT EXISTS idx_homework_comments_homework_student_teacher
ON public.homework_comments (homework_id, student_id, teacher_id);

