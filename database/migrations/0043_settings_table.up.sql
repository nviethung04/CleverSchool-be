CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    values JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX idx_settings_key_active ON settings(key) WHERE deleted_at IS NULL;
CREATE INDEX idx_settings_deleted_at ON settings(deleted_at);
