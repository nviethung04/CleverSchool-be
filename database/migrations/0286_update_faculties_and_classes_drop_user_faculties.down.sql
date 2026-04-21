-- +migrate Down
-- Revert các thay đổi ở 0286

-- Xóa school_id khỏi faculties
ALTER TABLE faculties
    DROP COLUMN IF EXISTS school_id;

-- Xóa faculty_id khỏi classes
ALTER TABLE classes
    DROP COLUMN IF EXISTS faculty_id;

-- Tạo lại bảng user_faculties và index như ban đầu
CREATE TABLE IF NOT EXISTS user_faculties (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    faculty_id INTEGER NOT NULL,
    CONSTRAINT uq_user_faculties UNIQUE (user_id, faculty_id)
);

CREATE INDEX IF NOT EXISTS idx_user_faculties_user_id ON user_faculties(user_id);
CREATE INDEX IF NOT EXISTS idx_user_faculties_faculty_id ON user_faculties(faculty_id);


