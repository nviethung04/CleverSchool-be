-- +migrate Up

CREATE TABLE IF NOT EXISTS assessment_criteria_group_ref_criteria (
    id                          BIGSERIAL PRIMARY KEY,
    assessment_criteria_group_id BIGINT NOT NULL,
    assessment_criteria_id      BIGINT NOT NULL,
    created_at                  TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    created_by                  BIGINT,
    updated_at                  TIMESTAMP WITHOUT TIME ZONE,
    updated_by                  BIGINT,
    deleted_at                  TIMESTAMP WITHOUT TIME ZONE,
    deleted_by                  BIGINT
);

CREATE INDEX IF NOT EXISTS idx_acgrc_group_id ON assessment_criteria_group_ref_criteria (assessment_criteria_group_id);
CREATE INDEX IF NOT EXISTS idx_acgrc_criteria_id ON assessment_criteria_group_ref_criteria (assessment_criteria_id);
CREATE INDEX IF NOT EXISTS idx_acgrc_deleted_at ON assessment_criteria_group_ref_criteria (deleted_at);
CREATE INDEX IF NOT EXISTS idx_acgrc_deleted_by ON assessment_criteria_group_ref_criteria (deleted_by);


