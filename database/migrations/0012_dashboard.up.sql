CREATE TABLE dashboard_report_schools (
    id BIGSERIAL PRIMARY KEY,
    school_id BIGINT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    total_students BIGINT DEFAULT 0,
    total_teachers BIGINT DEFAULT 0,
    active_students BIGINT DEFAULT 0,
    active_teachers BIGINT DEFAULT 0,
    students_completed_homework BIGINT DEFAULT 0,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_dashboard_report_schools_school_id ON dashboard_report_schools(school_id);
CREATE INDEX idx_dashboard_report_schools_dates ON dashboard_report_schools(start_date, end_date);

CREATE TABLE dashboard_report_courses (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    total_students BIGINT DEFAULT 0,
    total_teachers BIGINT DEFAULT 0,
    active_students BIGINT DEFAULT 0,
    active_teachers BIGINT DEFAULT 0,
    students_completed_homework BIGINT DEFAULT 0,
    teacher_ids TEXT DEFAULT '',
    teacher_infos JSONB DEFAULT '[]'::jsonb,
    total_homeworks BIGINT DEFAULT 0,
    assigned_homeworks BIGINT DEFAULT 0,
    completed_homeworks BIGINT DEFAULT 0,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (course_id, start_date, end_date)
);
CREATE INDEX idx_dashboard_report_courses_start_date ON dashboard_report_courses(start_date);
CREATE INDEX idx_dashboard_report_courses_end_date ON dashboard_report_courses(end_date);

-- SQL functions cho dashboard overview: xem migrations_legacy/0258 hoặc thêm migration 0021 khi bật analytics đầy đủ.
