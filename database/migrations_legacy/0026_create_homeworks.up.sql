-- +migrate Up

CREATE TABLE homeworks
(
    id         BIGSERIAL PRIMARY KEY,
    lesson_id  BIGINT,
    name       VARCHAR(100),
    status     SMALLINT,
    type       TEXT NOT NULL DEFAULT 'homework',
    created_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    FOREIGN KEY (lesson_id) REFERENCES lessons (id) ON DELETE CASCADE
);

CREATE INDEX idx_homeworks_lesson ON homeworks (lesson_id);
CREATE INDEX idx_homeworks_status ON homeworks (status);
CREATE INDEX idx_homeworks_type ON homeworks (type);