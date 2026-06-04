-- +migrate Up

CREATE TABLE lesson_plans
(
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(100),
    description   TEXT,
    status        SMALLINT,
    sort_position INT,
    total_time    BIGINT,
    type          TEXT NOT NULL DEFAULT 'lesson_plan',
    created_at    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT
);

CREATE INDEX idx_lesson_plans_status ON lesson_plans (status);
CREATE INDEX idx_lesson_plans_type ON lesson_plans (type);