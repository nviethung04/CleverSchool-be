-- Align lesson_plan_parts with models/lesson_plan_part.go (baseline 0007 was minimal).

ALTER TABLE lesson_plan_parts
    ADD COLUMN IF NOT EXISTS program_id BIGINT REFERENCES programs(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS course_id BIGINT,
    ADD COLUMN IF NOT EXISTS tag VARCHAR(255),
    ADD COLUMN IF NOT EXISTS time BIGINT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_classwork BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS file_type VARCHAR(100),
    ADD COLUMN IF NOT EXISTS link_info JSONB,
    ADD COLUMN IF NOT EXISTS link_type VARCHAR(100),
    ADD COLUMN IF NOT EXISTS guide_teacher TEXT,
    ADD COLUMN IF NOT EXISTS guide_student TEXT,
    ADD COLUMN IF NOT EXISTS file TEXT,
    ADD COLUMN IF NOT EXISTS max_score NUMERIC(10, 2) DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_lesson_plan_parts_program_id ON lesson_plan_parts(program_id);
CREATE INDEX IF NOT EXISTS idx_lesson_plan_parts_course_id ON lesson_plan_parts(course_id);
