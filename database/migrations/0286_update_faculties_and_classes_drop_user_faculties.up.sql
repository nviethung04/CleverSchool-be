-- +migrate Up
-- Xóa bảng user_faculties
DROP INDEX IF EXISTS idx_user_faculties_user_id;
DROP INDEX IF EXISTS idx_user_faculties_faculty_id;
DROP TABLE IF EXISTS user_faculties;

-- Thêm faculty_id vào bảng classes
ALTER TABLE classes
    ADD COLUMN IF NOT EXISTS faculty_id BIGINT;

-- Thêm school_id vào bảng faculties
ALTER TABLE faculties
    ADD COLUMN IF NOT EXISTS school_id BIGINT;


