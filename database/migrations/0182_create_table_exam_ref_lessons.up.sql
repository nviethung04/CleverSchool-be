CREATE TABLE exam_ref_lessons (
    id SERIAL PRIMARY KEY,
    lesson_id INTEGER NOT NULL,
    exam_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_exam_ref_lessons UNIQUE (lesson_id, exam_id)
);

CREATE INDEX idx_exam_ref_lessons_lesson_id ON exam_ref_lessons(lesson_id);
CREATE INDEX idx_exam_ref_lessons_homework_id ON exam_ref_lessons(exam_id);
