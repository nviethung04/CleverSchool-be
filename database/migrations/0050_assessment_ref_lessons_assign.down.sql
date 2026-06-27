DROP INDEX IF EXISTS assessment_ref_lessons_course_uq;
DROP INDEX IF EXISTS assessment_ref_lessons_program_uq;

ALTER TABLE assessment_ref_lessons
    DROP COLUMN IF EXISTS assigned_by,
    DROP COLUMN IF EXISTS assigned_at;

ALTER TABLE assessment_ref_lessons
    ADD CONSTRAINT assessment_ref_lessons_lesson_id_assessment_id_key UNIQUE (lesson_id, assessment_id);
