CREATE TABLE dashboard_report_schools (
    id BIGSERIAL PRIMARY KEY,
    school_id BIGINT NOT NULL,
    total_students BIGINT DEFAULT 0,
    total_teachers BIGINT DEFAULT 0,
    active_students BIGINT DEFAULT 0,
    active_teachers BIGINT DEFAULT 0,
    students_completed_homework BIGINT DEFAULT 0,
    start_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    end_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_dashboard_report_schools_school_id 
        FOREIGN KEY (school_id) REFERENCES schools(id)
);

-- Tạo index để tối ưu query
CREATE INDEX idx_dashboard_report_schools_school_id ON dashboard_report_schools(school_id);
CREATE INDEX idx_dashboard_report_schools_dates ON dashboard_report_schools(start_date, end_date);
CREATE INDEX idx_dashboard_report_schools_created_at ON dashboard_report_schools(created_at);
