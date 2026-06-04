-- Xóa 3 cột thống kê homework từ bảng dashboard_report_courses
ALTER TABLE dashboard_report_courses 
DROP COLUMN IF EXISTS total_homeworks,
DROP COLUMN IF EXISTS assigned_homeworks,
DROP COLUMN IF EXISTS completed_homeworks;

