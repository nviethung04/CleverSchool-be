-- +migrate Up

CREATE TABLE lesson_plan_parts_questions
(
    lesson_plan_parts_id BIGINT NOT NULL,
    questions_id         BIGINT NOT NULL,
    sort_position        INT,
    PRIMARY KEY (lesson_plan_parts_id, questions_id),
    FOREIGN KEY (lesson_plan_parts_id) REFERENCES lesson_plan_parts (id) ON DELETE CASCADE,
    FOREIGN KEY (questions_id) REFERENCES questions (id) ON DELETE CASCADE
);

CREATE INDEX idx_lesson_plan_parts_questions_part ON lesson_plan_parts_questions (lesson_plan_parts_id);
CREATE INDEX idx_lesson_plan_parts_questions_question ON lesson_plan_parts_questions (questions_id);