-- Bước 1: Thêm cột mới file_infos kiểu text[]
ALTER TABLE questions ADD COLUMN file_infos jsonb;

-- Bước 2: Cập nhật dữ liệu từ JSONB sang mảng text[]
UPDATE questions
SET file_infos = jsonb_build_array(file_info);
