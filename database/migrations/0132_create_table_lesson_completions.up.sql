CREATE TABLE lesson_completions (
    student_id BIGINT NOT NULL,
    lesson_id BIGINT NOT NULL,
    completed_at TIMESTAMP DEFAULT now(),

    UNIQUE (student_id, lesson_id)
);
