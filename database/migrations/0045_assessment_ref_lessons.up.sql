CREATE TABLE assessment_ref_lessons (
    id BIGSERIAL PRIMARY KEY,
    assessment_id BIGINT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    course_id BIGINT REFERENCES courses(id) ON DELETE CASCADE,
    UNIQUE (lesson_id, assessment_id)
);

CREATE INDEX idx_assessment_ref_lessons_lesson_id ON assessment_ref_lessons(lesson_id);
CREATE INDEX idx_assessment_ref_lessons_assessment_id ON assessment_ref_lessons(assessment_id);
