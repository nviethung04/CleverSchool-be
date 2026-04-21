-- +migrate Up

CREATE TABLE permissions
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    permission  VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(255),
    type        TEXT         NOT NULL DEFAULT 'permission',
    created_at  TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by  INT,
    updated_by  INT
);

CREATE INDEX idx_permissions_name ON permissions (name);
CREATE INDEX idx_permissions_permission ON permissions (permission);
CREATE INDEX idx_permissions_type ON permissions (type);