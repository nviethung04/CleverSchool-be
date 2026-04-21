-- Thêm 3 cột cho thống kê học sinh và homework vào bảng dashboard_report_courses
ALTER TABLE dashboard_report_courses 
ADD COLUMN students_completed_all_homeworks BIGINT DEFAULT 0,
ADD COLUMN students_doing_homeworks BIGINT DEFAULT 0,
ADD COLUMN students_not_started_any_homework BIGINT DEFAULT 0;

-- Thêm comment cho các cột
COMMENT ON COLUMN dashboard_report_courses.students_completed_all_homeworks IS 'Số lượng học sinh đã hoàn thành tất cả các bài homework được giao';
COMMENT ON COLUMN dashboard_report_courses.students_doing_homeworks IS 'Số lượng học sinh đang làm homework (tổng học sinh - hoàn thành tất cả - chưa làm gì)';
COMMENT ON COLUMN dashboard_report_courses.students_not_started_any_homework IS 'Số lượng học sinh chưa làm homework nào';

