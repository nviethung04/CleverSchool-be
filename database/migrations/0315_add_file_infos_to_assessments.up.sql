-- +migrate Up

ALTER TABLE assessments
    ADD COLUMN IF NOT EXISTS file_infos jsonb;

ALTER TABLE assessment_scores
    ADD COLUMN IF NOT EXISTS file_infos jsonb;

