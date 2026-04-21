CREATE TABLE IF NOT EXISTS activity_logs (
    id SERIAL PRIMARY KEY,
    user_id BIGINT,
    role_id INTEGER,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    client_ip VARCHAR(45),
    device VARCHAR(20),
    latency_ms INTEGER,
    query_params JSONB,
    request_body JSONB,
    response_body JSONB,
    attribute JSONB,
    session_id TEXT,
    agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION cleanup_old_activity_logs_table()
RETURNS TEXT AS $$
DECLARE
    table_exists BOOLEAN;
BEGIN
    -- Kiểm tra xem bảng cũ có tồn tại không
    SELECT EXISTS (
        SELECT FROM information_schema.tables
        WHERE table_schema = 'public'
        AND table_name = 'activity_logs'
    ) INTO table_exists;

    IF NOT table_exists THEN
        RETURN 'Table activity_logs does not exist';
    END IF;

    -- Xóa bảng cũ
    DROP TABLE activity_logs;

    RETURN 'Successfully dropped old activity_logs table';
END;
$$ LANGUAGE plpgsql;
