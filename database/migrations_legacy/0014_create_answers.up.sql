-- +migrate Up

CREATE TABLE answers
(
    id            BIGSERIAL PRIMARY KEY,
    question_id   INT       NOT NULL,
    content       TEXT,
    file_url      VARCHAR(255),
    is_correct    BOOLEAN            DEFAULT FALSE,
    kind          kind_enum NOT NULL,
    point         INT,
    sort_position INT,
    type          TEXT      NOT NULL DEFAULT 'answer',
    created_at    TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE
);

CREATE INDEX idx_answers_question ON answers (question_id);
CREATE INDEX idx_answers_is_correct ON answers (is_correct);
CREATE INDEX idx_answers_type ON answers (type);