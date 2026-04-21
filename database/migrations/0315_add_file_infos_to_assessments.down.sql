-- +migrate Down

ALTER TABLE assessment_scores
    DROP COLUMN IF EXISTS file_infos;

ALTER TABLE assessments
    DROP COLUMN IF EXISTS file_infos;

