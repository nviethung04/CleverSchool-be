-- +migrate Up

CREATE TABLE lesson_plan_parts
(
    id             BIGSERIAL PRIMARY KEY,
    lesson_plan_id BIGINT,
    title          TEXT,
    tag            VARCHAR(100),
    cover_image    TEXT,
    file           TEXT[],
    sort_position  SMALLINT,
    "time"         BIGINT,
    is_classwork   BOOLEAN NOT NULL DEFAULT FALSE,
    file_type      VARCHAR(100),
    link           TEXT,
    link_type      VARCHAR(100),
    guide_teacher  TEXT,
    guide_student  TEXT,
    type           TEXT    NOT NULL DEFAULT 'lesson_plan_part',
    created_at     TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    created_by     INT,
    updated_by     INT,
    FOREIGN KEY (lesson_plan_id) REFERENCES lesson_plans (id) ON DELETE CASCADE
);

CREATE INDEX idx_lesson_plan_parts_plan ON lesson_plan_parts (lesson_plan_id);
CREATE INDEX idx_lesson_plan_parts_is_classwork ON lesson_plan_parts (is_classwork);
CREATE INDEX idx_lesson_plan_parts_type ON lesson_plan_parts (type);