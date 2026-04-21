-- +migrate Up

CREATE TABLE IF NOT EXISTS assessment_criteria_groups (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    subject_id  BIGINT,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    created_by  BIGINT,
    updated_at  TIMESTAMP WITHOUT TIME ZONE,
    updated_by  BIGINT,
    deleted_at  TIMESTAMP WITHOUT TIME ZONE,
    deleted_by  BIGINT
);

CREATE INDEX IF NOT EXISTS idx_assessment_criteria_groups_subject_id ON assessment_criteria_groups (subject_id);
CREATE INDEX IF NOT EXISTS idx_assessment_criteria_groups_created_by ON assessment_criteria_groups (created_by);
CREATE INDEX IF NOT EXISTS idx_assessment_criteria_groups_deleted_at ON assessment_criteria_groups (deleted_at);
CREATE INDEX IF NOT EXISTS idx_assessment_criteria_groups_deleted_by ON assessment_criteria_groups (deleted_by);


