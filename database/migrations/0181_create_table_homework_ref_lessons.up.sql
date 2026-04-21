CREATE TABLE homework_ref_lessons (
    id SERIAL PRIMARY KEY,
    lesson_id INTEGER NOT NULL,
    homework_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_homework_ref_lessons UNIQUE (lesson_id, homework_id)
);

CREATE INDEX idx_homework_ref_lessons_lesson_id ON homework_ref_lessons(lesson_id);
CREATE INDEX idx_homework_ref_lessons_homework_id ON homework_ref_lessons(homework_id);
