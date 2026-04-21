-- +migrate Up

CREATE TABLE subjects
(
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL,
    status      BOOLEAN      NOT NULL DEFAULT FALSE,
    type        TEXT         NOT NULL DEFAULT 'subject',
    created_at  TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by  INT,
    updated_by  INT
);

CREATE INDEX idx_subjects_name ON subjects (name);
CREATE INDEX idx_subjects_status ON subjects (status);
CREATE INDEX idx_subjects_type ON subjects (type);