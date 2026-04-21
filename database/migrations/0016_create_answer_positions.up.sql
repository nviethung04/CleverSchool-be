-- +migrate Up

CREATE TABLE answer_positions
(
    id               BIGSERIAL PRIMARY KEY,
    question_id      INT       NOT NULL,
    content          TEXT,
    file_url         VARCHAR(255),
    is_correct       BOOLEAN            DEFAULT FALSE,
    kind             kind_enum NOT NULL,
    point            INT,
    sort_position    INT,
    correct_position INT,
    type             TEXT      NOT NULL DEFAULT 'answer_position',
    created_at       TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    created_by       INT,
    updated_by       INT,
    FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE
);

CREATE INDEX idx_answer_positions_question ON answer_positions (question_id);
CREATE INDEX idx_answer_positions_is_correct ON answer_positions (is_correct);
CREATE INDEX idx_answer_positions_type ON answer_positions (type);