-- Thay đổi kiểu dữ liệu của created_by và updated_by từ INT thành BIGINT
ALTER TABLE roles 
    ALTER COLUMN created_by TYPE BIGINT,
    ALTER COLUMN updated_by TYPE BIGINT;

-- Thêm các cột mới cho soft delete
ALTER TABLE roles
    ADD COLUMN deleted_at TIMESTAMP,
    ADD COLUMN deleted_by BIGINT;
