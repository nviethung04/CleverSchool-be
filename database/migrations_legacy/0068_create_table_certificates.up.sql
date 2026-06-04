CREATE TABLE certificates (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    issued_by VARCHAR(100),
    issued_date DATE,
    expiry_date DATE,
    certificate_code VARCHAR(50),
    status VARCHAR(20),
    description TEXT,
    file_url TEXT,

    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    created_by INT,
    updated_by INT,
    deleted_by INT
);
