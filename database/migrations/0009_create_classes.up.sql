-- +migrate Up

CREATE TABLE classes
(
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    status        SMALLINT     NOT NULL DEFAULT 0,
    sort_position INT,
    type          TEXT         NOT NULL DEFAULT 'class',
    created_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT
);

CREATE INDEX idx_classes_name ON classes (name);
CREATE INDEX idx_classes_status ON classes (status);
CREATE INDEX idx_classes_type ON classes (type);