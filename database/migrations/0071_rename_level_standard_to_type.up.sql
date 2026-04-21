-- Đổi tên cột trong bảng level_tests
ALTER TABLE level_tests RENAME COLUMN level_standard_id TO level_type_id;

-- Đổi tên cột trong bảng levels
ALTER TABLE levels RENAME COLUMN level_standard_id TO level_type_id;

-- Đổi tên bảng level_standards thành level_types
ALTER TABLE level_standards RENAME TO level_types;
