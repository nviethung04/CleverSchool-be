-- +migrate Up

CREATE TABLE group_answers
(
    id            BIGSERIAL PRIMARY KEY,
    content       TEXT,
    file_url      VARCHAR(255),
    kind          kind_enum NOT NULL,
    sort_position INT,
    type          TEXT      NOT NULL DEFAULT 'group_answer',
    created_at    TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT
);

CREATE INDEX idx_group_answers_type ON group_answers (type);