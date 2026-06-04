-- +migrate Up

CREATE TABLE questions_exams
(
    id            BIGSERIAL PRIMARY KEY,
    questions_id  BIGINT NOT NULL,
    exams_id      BIGINT NOT NULL,
    max_point     NUMERIC,
    time_limit    BIGINT,
    sort_position SMALLINT,
    type          TEXT   NOT NULL DEFAULT 'question_exam',
    created_at    TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP       DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    FOREIGN KEY (questions_id) REFERENCES questions (id) ON DELETE CASCADE,
    FOREIGN KEY (exams_id) REFERENCES exams (id) ON DELETE CASCADE
);

CREATE INDEX idx_questions_exams_question ON questions_exams (questions_id);
CREATE INDEX idx_questions_exams_exam ON questions_exams (exams_id);
CREATE INDEX idx_questions_exams_type ON questions_exams (type);