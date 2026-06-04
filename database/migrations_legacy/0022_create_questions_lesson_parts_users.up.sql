-- +migrate Up

CREATE TABLE questions_lesson_parts_users
(
    questions_lesson_parts_id BIGINT NOT NULL,
    users_id                  INT    NOT NULL,
    answer                    TEXT[],
    is_correct                BOOLEAN,
    point                     NUMERIC,
    PRIMARY KEY (questions_lesson_parts_id, users_id),
    FOREIGN KEY (questions_lesson_parts_id) REFERENCES lesson_plan_parts (id) ON DELETE CASCADE,
    FOREIGN KEY (users_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_questions_lesson_parts_users_part ON questions_lesson_parts_users (questions_lesson_parts_id);
CREATE INDEX idx_questions_lesson_parts_users_user ON questions_lesson_parts_users (users_id);