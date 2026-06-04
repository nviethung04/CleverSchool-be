CREATE TABLE dashboard_report_courses (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL,
    total_students BIGINT DEFAULT 0,
    total_teachers BIGINT DEFAULT 0,
    active_students BIGINT DEFAULT 0,
    active_teachers BIGINT DEFAULT 0,
    students_completed_homework BIGINT DEFAULT 0,
    teacher_ids TEXT DEFAULT '', -- Danh sách ID các teacher thuộc khóa, phân cách bằng dấu phẩy
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_course
        FOREIGN KEY(course_id)
        REFERENCES courses(id)
        ON DELETE CASCADE
);

-- Indexes
CREATE UNIQUE INDEX idx_dashboard_report_courses_course_date_range ON dashboard_report_courses(course_id, start_date, end_date);
CREATE INDEX idx_dashboard_report_courses_start_date ON dashboard_report_courses(start_date);
CREATE INDEX idx_dashboard_report_courses_end_date ON dashboard_report_courses(end_date);
CREATE INDEX idx_dashboard_report_courses_created_at ON dashboard_report_courses(created_at);
