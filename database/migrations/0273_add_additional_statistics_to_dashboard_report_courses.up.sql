-- Thêm 2 cột thống kê mới vào bảng dashboard_report_courses
ALTER TABLE dashboard_report_courses 
ADD COLUMN student_active_not_started_any_homework BIGINT DEFAULT 0,
ADD COLUMN homework_over_50_percent_student_complete BIGINT DEFAULT 0;

-- Thêm comment cho các cột
COMMENT ON COLUMN dashboard_report_courses.student_active_not_started_any_homework IS 'Số lượng học sinh active nhưng chưa làm homework nào';
COMMENT ON COLUMN dashboard_report_courses.homework_over_50_percent_student_complete IS 'Số lượng homework có trên 50% học sinh hoàn thành';

