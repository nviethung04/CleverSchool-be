CREATE TABLE cloned_questions (
    id SERIAL PRIMARY KEY,
    assignment_id INTEGER NOT NULL,
    assignment_type TEXT NOT NULL CHECK (assignment_type IN ('homework', 'exam', 'lesson_plan_part', 'level_test')),
    questions JSONB NOT NULL,

    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

CREATE INDEX idx_cloned_questions_assignment ON cloned_questions(assignment_type, assignment_id);
