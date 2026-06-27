-- Giao bài đánh giá theo khóa (tương tự homework_ref_lessons.assigned_by).

ALTER TABLE assessment_ref_lessons
    ADD COLUMN IF NOT EXISTS assigned_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS assigned_by BIGINT REFERENCES users(id);

ALTER TABLE assessment_ref_lessons
    DROP CONSTRAINT IF EXISTS assessment_ref_lessons_lesson_id_assessment_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS assessment_ref_lessons_program_uq
    ON assessment_ref_lessons (lesson_id, assessment_id) WHERE course_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS assessment_ref_lessons_course_uq
    ON assessment_ref_lessons (lesson_id, assessment_id, course_id) WHERE course_id IS NOT NULL;
