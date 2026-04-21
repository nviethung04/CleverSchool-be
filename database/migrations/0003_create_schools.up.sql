-- +migrate Up

CREATE TABLE schools
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100),
    type       TEXT NOT NULL DEFAULT 'school',
    created_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT
);

CREATE INDEX idx_schools_name ON schools (name);
CREATE INDEX idx_schools_type ON schools (type);