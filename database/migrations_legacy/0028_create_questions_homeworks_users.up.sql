-- +migrate Up

CREATE TABLE questions_homeworks_users
(
    questions_homeworks_id BIGINT NOT NULL,
    users_id               INT    NOT NULL,
    answer                 TEXT,
    is_correct             BOOLEAN,
    file                   TEXT,
    point                  NUMERIC,
    "time"                 BIGINT,
    has_point              BOOLEAN,
    scoring_by             INT,
    PRIMARY KEY (questions_homeworks_id, users_id),
    FOREIGN KEY (questions_homeworks_id) REFERENCES questions_homeworks (id) ON DELETE CASCADE,
    FOREIGN KEY (users_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (scoring_by) REFERENCES users (id) ON DELETE SET NULL
);

CREATE INDEX idx_questions_homeworks_users_homework ON questions_homeworks_users (questions_homeworks_id);
CREATE INDEX idx_questions_homeworks_users_user ON questions_homeworks_users (users_id);