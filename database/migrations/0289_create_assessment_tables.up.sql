DO $$ BEGIN
    CREATE TYPE public.assessment_type_enum AS ENUM
        ('mini_test', 'final_exam');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;


CREATE TABLE IF NOT EXISTS assessment_criteria (
    id bigserial,
    subject_id BIGINT,
    name TEXT,
    description TEXT,
    has_subcriteria BOOLEAN,
    max_score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT,
    CONSTRAINT assessment_criteria_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_assessment_criteria_created_by ON assessment_criteria (created_by);
CREATE INDEX IF NOT EXISTS idx_assessment_criteria_deleted_at ON assessment_criteria (deleted_at);
CREATE INDEX IF NOT EXISTS idx_assessment_criteria_deleted_by ON assessment_criteria (deleted_by);
CREATE INDEX IF NOT EXISTS idx_assessment_criteria_subject_id ON assessment_criteria (subject_id);
CREATE INDEX IF NOT EXISTS idx_assessment_criteria_updated_by ON assessment_criteria (updated_by);

CREATE TABLE IF NOT EXISTS assessment_subcriteria (
    id bigserial,
    criterion_id BIGINT,
    name TEXT,
    description TEXT,
    max_score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT,
    CONSTRAINT assessment_subcriteria_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_assessment_subcriteria_created_by ON assessment_subcriteria (created_by);
CREATE INDEX IF NOT EXISTS idx_assessment_subcriteria_criterion_id ON assessment_subcriteria (criterion_id);
CREATE INDEX IF NOT EXISTS idx_assessment_subcriteria_deleted_at ON assessment_subcriteria (deleted_at);
CREATE INDEX IF NOT EXISTS idx_assessment_subcriteria_deleted_by ON assessment_subcriteria (deleted_by);
CREATE INDEX IF NOT EXISTS idx_assessment_subcriteria_updated_by ON assessment_subcriteria (updated_by);

CREATE TABLE IF NOT EXISTS assessments (
    id bigserial,
    name TEXT,
    description TEXT,
    type assessment_type_enum,
    program_id BIGINT,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    update_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT,
    CONSTRAINT assessments_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_assessments_created_by ON assessments (created_by);
CREATE INDEX IF NOT EXISTS idx_assessments_deleted_at ON assessments (deleted_at);
CREATE INDEX IF NOT EXISTS idx_assessments_deleted_by ON assessments (deleted_by);
CREATE INDEX IF NOT EXISTS idx_assessments_program_id ON assessments (program_id);
CREATE INDEX IF NOT EXISTS idx_assessments_updated_by ON assessments (updated_by);

CREATE TABLE IF NOT EXISTS assessment_ref_lessons (
    id bigserial,
    assessment_id INTEGER NOT NULL,
    lesson_id INTEGER NOT NULL,
    course_id BIGINT,
    assigned_by BIGINT,
    assigned_at TIMESTAMP WITHOUT TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT assessment_ref_lessons_pkey PRIMARY KEY (id),
    CONSTRAINT assessment_ref_lessons_unique UNIQUE (assessment_id, lesson_id, course_id)
);

CREATE INDEX IF NOT EXISTS idx_assessment_ref_lessons_assessment_id ON assessment_ref_lessons (assessment_id);
CREATE INDEX IF NOT EXISTS idx_assessment_ref_lessons_assigned_at ON assessment_ref_lessons (assigned_at);
CREATE INDEX IF NOT EXISTS idx_assessment_ref_lessons_assigned_by ON assessment_ref_lessons (assigned_by);
CREATE INDEX IF NOT EXISTS idx_assessment_ref_lessons_course_id ON assessment_ref_lessons (course_id);
CREATE INDEX IF NOT EXISTS idx_assessment_ref_lessons_lesson_id ON assessment_ref_lessons (lesson_id);

CREATE TABLE IF NOT EXISTS assessment_scores (
    id bigserial,
    assessment_id BIGINT,
    student_id BIGINT,
    lesson_id BIGINT,
    course_id BIGINT,
    total_score NUMERIC(8,2),
    status_scored INTEGER,
    is_late BOOLEAN,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT,
    CONSTRAINT assessment_scores_pkey PRIMARY KEY (id)
);

COMMENT ON COLUMN assessment_scores.status_scored IS '0: chưa chấm

1: đang chấm (chưa xong)

2: đã chấm xong';

CREATE INDEX IF NOT EXISTS idx_assessment_scores_assessment_id ON assessment_scores (assessment_id);
CREATE INDEX IF NOT EXISTS idx_assessment_scores_course_id ON assessment_scores (course_id);
CREATE INDEX IF NOT EXISTS idx_assessment_scores_created_by ON assessment_scores (created_by);
CREATE INDEX IF NOT EXISTS idx_assessment_scores_deleted_at ON assessment_scores (deleted_at);
CREATE INDEX IF NOT EXISTS idx_assessment_scores_deleted_by ON assessment_scores (deleted_by);
CREATE INDEX IF NOT EXISTS idx_assessment_scores_lesson_id ON assessment_scores (lesson_id);
CREATE INDEX IF NOT EXISTS idx_assessment_scores_student_id ON assessment_scores (student_id);
CREATE INDEX IF NOT EXISTS idx_assessment_scores_updated_by ON assessment_scores (updated_by);

CREATE TABLE IF NOT EXISTS assessment_score_details (
    id bigserial,
    assessment_score_id BIGINT,
    criterion_id BIGINT,
    subcriterion_id BIGINT,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT,
    CONSTRAINT assessment_score_details_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_assessment_score_details_assessment_score_id ON assessment_score_details (assessment_score_id);
CREATE INDEX IF NOT EXISTS idx_assessment_score_details_created_by ON assessment_score_details (created_by);
CREATE INDEX IF NOT EXISTS idx_assessment_score_details_criterion_id ON assessment_score_details (criterion_id);
CREATE INDEX IF NOT EXISTS idx_assessment_score_details_deleted_at ON assessment_score_details (deleted_at);
CREATE INDEX IF NOT EXISTS idx_assessment_score_details_deleted_by ON assessment_score_details (deleted_by);
CREATE INDEX IF NOT EXISTS idx_assessment_score_details_subcriterion_id ON assessment_score_details (subcriterion_id);
CREATE INDEX IF NOT EXISTS idx_assessment_score_details_updated_by ON assessment_score_details (updated_by);

