-- Align lesson_plans with models/lesson_plan.go (baseline 0007 was minimal).

ALTER TABLE lesson_plans
    ADD COLUMN IF NOT EXISTS program_id BIGINT REFERENCES programs(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS object_title VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS cover_image_info JSONB,
    ADD COLUMN IF NOT EXISTS sort_position INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_time INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS views INT DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_lesson_plans_program_id ON lesson_plans(program_id);
CREATE INDEX IF NOT EXISTS idx_lesson_plans_sort_position ON lesson_plans(sort_position);
