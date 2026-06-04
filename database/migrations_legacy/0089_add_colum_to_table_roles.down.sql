-- Rollback: Xóa các cột soft delete
ALTER TABLE roles
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_by;

-- Rollback: Thay đổi lại kiểu dữ liệu về INT
ALTER TABLE roles 
    ALTER COLUMN created_by TYPE INT,
    ALTER COLUMN updated_by TYPE INT;
