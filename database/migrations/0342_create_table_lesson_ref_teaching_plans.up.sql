CREATE TABLE lesson_ref_teaching_plans (
    id SERIAL PRIMARY KEY,
    lesson_id INTEGER NOT NULL,
    teaching_plan_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_lesson_ref_teaching_plans UNIQUE (lesson_id, teaching_plan_id)
);

CREATE INDEX idx_lesson_ref_teaching_plans_lesson_id ON lesson_ref_teaching_plans(lesson_id);
CREATE INDEX idx_lesson_ref_teaching_plans_teaching_plan_id ON lesson_ref_teaching_plans(teaching_plan_id);
