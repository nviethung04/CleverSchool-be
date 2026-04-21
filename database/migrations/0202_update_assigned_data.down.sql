-- Xóa các cột vừa thêm ở bảng trung gian
ALTER TABLE homework_ref_lessons
    DROP COLUMN assigned_by,
    DROP COLUMN assigned_at;

ALTER TABLE exam_ref_lessons
    DROP COLUMN assigned_by,
    DROP COLUMN assigned_at;

ALTER TABLE exercise_ref_lessons
    DROP COLUMN assigned_by,
    DROP COLUMN assigned_at;

-- Thêm lại các cột đã xóa ở bảng chính
ALTER TABLE homeworks
    ADD COLUMN assigned_by BIGINT,
    ADD COLUMN assigned_at TIMESTAMP,
    ADD COLUMN is_assigned BOOLEAN DEFAULT FALSE,
    ADD COLUMN lesson_id BIGINT;

ALTER TABLE exams
    ADD COLUMN assigned_by BIGINT,
    ADD COLUMN assigned_at TIMESTAMP,
    ADD COLUMN is_assigned BOOLEAN DEFAULT FALSE,
    ADD COLUMN lesson_id BIGINT;

ALTER TABLE exercises
    ADD COLUMN assigned_by BIGINT,
    ADD COLUMN assigned_at TIMESTAMP,
    ADD COLUMN is_assigned BOOLEAN DEFAULT FALSE,
    ADD COLUMN lesson_id BIGINT;
