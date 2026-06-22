ALTER TABLE assessments DROP CONSTRAINT IF EXISTS fk_assessments_study_report_criteria;
ALTER TABLE assessments DROP CONSTRAINT IF EXISTS fk_assessments_criteria_group;

DROP TABLE IF EXISTS study_report_publishes;
DROP TABLE IF EXISTS study_reports;
DROP TABLE IF EXISTS assessment_publishes;
DROP TABLE IF EXISTS assessment_score_details;
DROP TABLE IF EXISTS assessment_scores;
DROP TABLE IF EXISTS study_report_criterias;
DROP TABLE IF EXISTS assessment_subcriteria;
DROP TABLE IF EXISTS assessment_criteria;
DROP TABLE IF EXISTS assessment_criteria_groups;
