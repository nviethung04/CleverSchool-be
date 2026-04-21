ALTER TABLE study_reports
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS type_value,
    ADD COLUMN assessment_score_id BIGINT,
    ADD COLUMN assessment_id BIGINT;

CREATE INDEX idx_study_reports_assessment_score_id ON study_reports(assessment_score_id);
CREATE INDEX idx_study_reports_assessment_id ON study_reports(assessment_id);
