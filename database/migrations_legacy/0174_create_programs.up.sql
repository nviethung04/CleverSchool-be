-- Tạo bảng programs
CREATE TABLE programs
(
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    status        BOOLEAN DEFAULT TRUE,
    subject_id    BIGINT NULL,
    image_info    JSONB,
    target        TEXT,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    deleted_by    INT
);
