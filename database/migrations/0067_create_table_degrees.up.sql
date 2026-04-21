CREATE TABLE degrees (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    major VARCHAR(100),
    institution VARCHAR(100),
    graduation_year INT,
    degree_level VARCHAR(50),
    received_date DATE,
    note TEXT,

    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    created_by INT,
    updated_by INT,
    deleted_by INT
);
