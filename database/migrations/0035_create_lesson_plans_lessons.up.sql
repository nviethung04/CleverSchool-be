-- +migrate Up

CREATE TABLE IF NOT EXISTS public.lesson_plans_lessons
(
    lesson_plans_id BIGSERIAL NOT NULL,
    lessons_id       SERIAL     NOT NULL,
    PRIMARY KEY (lesson_plans_id, lessons_id),
    FOREIGN KEY (lesson_plans_id) REFERENCES lesson_plans (id) ON DELETE CASCADE,
    FOREIGN KEY (lessons_id) REFERENCES lessons (id) ON DELETE CASCADE
    );

CREATE INDEX idx_lesson_plans_lessons_lesson_plans ON lesson_plans_lessons (lesson_plans_id);
CREATE INDEX idx_lesson_plans_lessons_lessons ON lesson_plans_lessons (lessons_id);
