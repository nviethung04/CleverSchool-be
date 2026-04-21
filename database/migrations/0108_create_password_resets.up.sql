-- +migrate Up

CREATE TABLE password_resets (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    email VARCHAR(100) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    status BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_password_resets_user_id ON password_resets (user_id);
CREATE INDEX idx_password_resets_email ON password_resets (email);
CREATE INDEX idx_password_resets_token ON password_resets (token);
CREATE INDEX idx_password_resets_status ON password_resets (status);
CREATE INDEX idx_password_resets_expires_at ON password_resets (expires_at);
