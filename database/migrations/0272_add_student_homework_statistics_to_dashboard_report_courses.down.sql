-- Xóa 3 cột thống kê học sinh và homework từ bảng dashboard_report_courses
ALTER TABLE dashboard_report_courses 
DROP COLUMN IF EXISTS students_completed_all_homeworks,
DROP COLUMN IF EXISTS students_doing_homeworks,
DROP COLUMN IF EXISTS students_not_started_any_homework;

