-- +migrate Up

CREATE TABLE questions_homeworks
(
    id            BIGSERIAL PRIMARY KEY,
    homeworks_id  BIGINT NOT NULL,
    questions_id  BIGINT NOT NULL,
    max_point     NUMERIC,
    sort_position SMALLINT,
    type          TEXT   NOT NULL DEFAULT 'question_homework',
    created_at    TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    FOREIGN KEY (homeworks_id) REFERENCES homeworks (id) ON DELETE CASCADE,
    FOREIGN KEY (questions_id) REFERENCES questions (id) ON DELETE CASCADE
);

CREATE INDEX idx_questions_homeworks_homework ON questions_homeworks (homeworks_id);
CREATE INDEX idx_questions_homeworks_question ON questions_homeworks (questions_id);
CREATE INDEX idx_questions_homeworks_type ON questions_homeworks (type);