DROP INDEX IF EXISTS idx_study_reports_assessment_score_id;
DROP INDEX IF EXISTS idx_study_reports_assessment_id;

ALTER TABLE study_reports
    DROP COLUMN IF EXISTS assessment_score_id,
    DROP COLUMN IF EXISTS assessment_id;

ALTER TABLE study_reports
    ADD COLUMN type VARCHAR(50) NOT NULL CHECK (type IN ('month', 'semester')),
    ADD COLUMN type_value TEXT;
