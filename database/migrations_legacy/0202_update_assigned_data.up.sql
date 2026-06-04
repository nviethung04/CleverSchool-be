ALTER TABLE homework_ref_lessons
    ADD COLUMN assigned_by BIGINT,
    ADD COLUMN assigned_at TIMESTAMP;

ALTER TABLE exam_ref_lessons
    ADD COLUMN assigned_by BIGINT,
    ADD COLUMN assigned_at TIMESTAMP;

ALTER TABLE exercise_ref_lessons
    ADD COLUMN assigned_by BIGINT,
    ADD COLUMN assigned_at TIMESTAMP;

ALTER TABLE homeworks DROP COLUMN assigned_by, DROP COLUMN assigned_at, DROP COLUMN is_assigned, DROP COLUMN lesson_id;
ALTER TABLE exams DROP COLUMN assigned_by, DROP COLUMN assigned_at, DROP COLUMN is_assigned, DROP COLUMN lesson_id;
ALTER TABLE exercises DROP COLUMN assigned_by, DROP COLUMN assigned_at, DROP COLUMN is_assigned, DROP COLUMN lesson_id;
