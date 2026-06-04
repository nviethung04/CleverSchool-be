-- Thêm cột teacher_infos kiểu JSONB vào bảng dashboard_report_courses
ALTER TABLE dashboard_report_courses 
ADD COLUMN teacher_infos JSONB DEFAULT '[]'::jsonb;

-- Thêm comment cho cột
COMMENT ON COLUMN dashboard_report_courses.teacher_infos IS 'JSON array chứa thông tin chi tiết của các teachers: [{"id": 1, "username": "teacher1", "name": "Teacher Name"}]';
