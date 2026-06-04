-- +migrate Up

CREATE TABLE exams
(
    id         BIGSERIAL PRIMARY KEY,
    lesson_id  BIGINT,
    name       VARCHAR(100),
    status     SMALLINT,
    time_limit BIGINT,
    type       TEXT NOT NULL DEFAULT 'exam',
    created_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    FOREIGN KEY (lesson_id) REFERENCES lessons (id) ON DELETE CASCADE
);

CREATE INDEX idx_exams_lesson ON exams (lesson_id);
CREATE INDEX idx_exams_status ON exams (status);
CREATE INDEX idx_exams_type ON exams (type);