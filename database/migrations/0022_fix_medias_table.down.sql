DROP TABLE IF EXISTS medias CASCADE;

CREATE TABLE medias (
    id BIGSERIAL PRIMARY KEY,
    file_name VARCHAR(500),
    file_path TEXT,
    file_type VARCHAR(100),
    file_size BIGINT,
    disk VARCHAR(50),
    static_url TEXT,
    type VARCHAR(50) DEFAULT 'media',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
