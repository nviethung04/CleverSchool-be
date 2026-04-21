-- Thêm cột is_assigned vào bảng homeworks
ALTER TABLE homeworks
    ADD COLUMN is_assigned BOOLEAN DEFAULT FALSE;

-- Thêm cột is_assigned vào bảng exams
ALTER TABLE exams
    ADD COLUMN is_assigned BOOLEAN DEFAULT FALSE;
