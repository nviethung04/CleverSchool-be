CREATE TABLE lesson_studying (
    student_id BIGINT NOT NULL,
    lesson_id BIGINT NOT NULL,
    studying_at TIMESTAMP DEFAULT now(),

    UNIQUE (student_id, lesson_id)
);
