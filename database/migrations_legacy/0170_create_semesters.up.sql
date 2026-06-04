-- Create semesters table
CREATE TABLE IF NOT EXISTS semesters (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    begin_date TIMESTAMP NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    previous_semester_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

-- Create course_ref_semester table for many-to-many relationship
CREATE TABLE IF NOT EXISTS course_ref_semesters (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL,
    semester_id BIGINT NOT NULL,
    UNIQUE(course_id, semester_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_course_ref_semesters_course_id ON course_ref_semesters(course_id);
CREATE INDEX IF NOT EXISTS idx_course_ref_semesters_semester_id ON course_ref_semesters(semester_id);
