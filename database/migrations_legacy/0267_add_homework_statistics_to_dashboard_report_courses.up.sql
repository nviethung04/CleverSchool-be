-- Thêm 3 cột cho thống kê homework vào bảng dashboard_report_courses
ALTER TABLE dashboard_report_courses 
ADD COLUMN total_homeworks BIGINT DEFAULT 0,
ADD COLUMN assigned_homeworks BIGINT DEFAULT 0,
ADD COLUMN completed_homeworks BIGINT DEFAULT 0;

-- Thêm comment cho các cột
COMMENT ON COLUMN dashboard_report_courses.total_homeworks IS 'Tổng số homework được tạo trong khoảng thời gian';
COMMENT ON COLUMN dashboard_report_courses.assigned_homeworks IS 'Số homework đã được giao (assigned_at IS NOT NULL)';
COMMENT ON COLUMN dashboard_report_courses.completed_homeworks IS 'Số homework đã hoàn thành (có ít nhất 1 user hoàn thành)';

