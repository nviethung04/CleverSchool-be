CREATE TABLE activity_logs (
    id SERIAL PRIMARY KEY,
    user_id BIGINT,
    role_id INTEGER,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    client_ip VARCHAR(45),
    latency_ms INTEGER,
    query_params TEXT,
    request_body TEXT,
    response_body TEXT,
    attribute TEXT,
    created_at TIMESTAMP DEFAULT NOW() 
);
