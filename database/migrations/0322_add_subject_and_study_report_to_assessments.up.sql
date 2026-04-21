ALTER TABLE assessments
    ADD COLUMN IF NOT EXISTS study_report_id BIGINT,
    ADD COLUMN IF NOT EXISTS subject_id BIGINT;

