-- Xóa 2 cột thống kê mới từ bảng dashboard_report_courses
ALTER TABLE dashboard_report_courses 
DROP COLUMN IF EXISTS student_active_not_started_any_homework,
DROP COLUMN IF EXISTS homework_over_50_percent_student_complete;

