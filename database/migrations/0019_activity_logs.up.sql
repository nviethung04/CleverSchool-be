CREATE TABLE activity_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    role_id BIGINT,
    school_id BIGINT,
    course_id BIGINT,
    action VARCHAR(255),
    resource VARCHAR(255),
    resource_id BIGINT,
    ip_address VARCHAR(100),
    user_agent TEXT,
    request_data JSONB,
    response_status INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at);
CREATE INDEX idx_activity_logs_school_id ON activity_logs(school_id);
