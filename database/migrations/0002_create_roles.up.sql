-- +migrate Up

CREATE TABLE roles
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    type       TEXT         NOT NULL DEFAULT 'role',
    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT
);

CREATE INDEX idx_roles_name ON roles (name);
CREATE INDEX idx_roles_type ON roles (type);