-- +migrate Up

CREATE TABLE questions_exams_users
(
    questions_exams_id BIGINT NOT NULL,
    users_id           INT    NOT NULL,
    file               TEXT,
    answer             TEXT,
    is_correct         BOOLEAN,
    point              NUMERIC,
    "time"             BIGINT,
    has_point          BOOLEAN,
    scoring_by         INT,
    PRIMARY KEY (questions_exams_id, users_id),
    FOREIGN KEY (questions_exams_id) REFERENCES questions_exams (id) ON DELETE CASCADE,
    FOREIGN KEY (users_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (scoring_by) REFERENCES users (id) ON DELETE SET NULL
);

CREATE INDEX idx_questions_exams_users_exam ON questions_exams_users (questions_exams_id);
CREATE INDEX idx_questions_exams_users_user ON questions_exams_users (users_id);