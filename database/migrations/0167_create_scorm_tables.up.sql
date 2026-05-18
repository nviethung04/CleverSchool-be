-- Migration: Create SCORM tables
-- Version: 001
-- Description: Create SCORM tables for learning management system


-- Drop old tables if exist
DROP TABLE IF EXISTS scorm_activities CASCADE;
DROP TABLE IF EXISTS scorm_attempts CASCADE;
DROP TABLE IF EXISTS scorm_cmi CASCADE;
DROP TABLE IF EXISTS scorm_sessions CASCADE;

-- Create scorm_activities table
CREATE TABLE IF NOT EXISTS scorm_activities (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(16) NOT NULL,
    launch_url TEXT NOT NULL,
    mastery_score DECIMAL(5,2),
    max_time_allowed VARCHAR(32),
    time_limit_action VARCHAR(32),
    launch_data TEXT,
    data_from_Clever School TEXT,
    status VARCHAR(32) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create scorm_attempts table
CREATE TABLE IF NOT EXISTS scorm_attempts (
    id VARCHAR(64) PRIMARY KEY,
    activity_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    version VARCHAR(16) NOT NULL,
    status VARCHAR(32) DEFAULT 'incomplete',
    score_raw DECIMAL(5,2),
    score_min DECIMAL(5,2),
    score_max DECIMAL(5,2),
    total_time VARCHAR(32),
    success_status VARCHAR(32),
    completion_status VARCHAR(32),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create scorm_cmi table
CREATE TABLE IF NOT EXISTS scorm_cmi (
    id BIGSERIAL PRIMARY KEY,
    attempt_id VARCHAR(64) NOT NULL,
    element VARCHAR(255) NOT NULL,
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create scorm_sessions table
CREATE TABLE IF NOT EXISTS scorm_sessions (
    id BIGSERIAL PRIMARY KEY,
    attempt_id VARCHAR(64) NOT NULL,
    session_id VARCHAR(64) NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_scorm_activities_status ON scorm_activities(status);
CREATE INDEX IF NOT EXISTS idx_scorm_activities_version ON scorm_activities(version);

CREATE INDEX IF NOT EXISTS idx_scorm_attempts_activity_id ON scorm_attempts(activity_id);
CREATE INDEX IF NOT EXISTS idx_scorm_attempts_user_id ON scorm_attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_scorm_attempts_status ON scorm_attempts(status);
CREATE INDEX IF NOT EXISTS idx_scorm_attempts_version ON scorm_attempts(version);

CREATE INDEX IF NOT EXISTS idx_scorm_cmi_attempt_id ON scorm_cmi(attempt_id);
CREATE INDEX IF NOT EXISTS idx_scorm_cmi_element ON scorm_cmi(element);
CREATE UNIQUE INDEX IF NOT EXISTS idx_scorm_cmi_attempt_element ON scorm_cmi(attempt_id, element);

CREATE INDEX IF NOT EXISTS idx_scorm_sessions_attempt_id ON scorm_sessions(attempt_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_scorm_sessions_session_id ON scorm_sessions(session_id);

-- Add foreign key constraints
ALTER TABLE scorm_attempts
ADD CONSTRAINT fk_scorm_attempts_activity_id
FOREIGN KEY (activity_id) REFERENCES scorm_activities(id) ON DELETE CASCADE;

ALTER TABLE scorm_cmi
ADD CONSTRAINT fk_scorm_cmi_attempt_id
FOREIGN KEY (attempt_id) REFERENCES scorm_attempts(id) ON DELETE CASCADE;

ALTER TABLE scorm_sessions
ADD CONSTRAINT fk_scorm_sessions_attempt_id
FOREIGN KEY (attempt_id) REFERENCES scorm_attempts(id) ON DELETE CASCADE;

-- Insert sample data for testing
INSERT INTO scorm_activities (title, description, version, launch_url, status) VALUES
('Introduction to Go Programming', 'Learn the basics of Go programming language', '1.2', '/scorm/go-basics/index.html', 'active'),
('Advanced Go Concepts', 'Advanced topics in Go programming', '2004', '/scorm/go-advanced/index.html', 'active'),
('Go Web Development', 'Building web applications with Go', '1.2', '/scorm/go-web/index.html', 'active')
ON CONFLICT DO NOTHING;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables
CREATE TRIGGER update_scorm_activities_updated_at
    BEFORE UPDATE ON scorm_activities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_scorm_attempts_updated_at
    BEFORE UPDATE ON scorm_attempts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_scorm_cmi_updated_at
    BEFORE UPDATE ON scorm_cmi
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_scorm_sessions_updated_at
    BEFORE UPDATE ON scorm_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
