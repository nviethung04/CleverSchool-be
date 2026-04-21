DROP INDEX IF EXISTS uq_study_report_ref_skill_types_report_skill;
DROP INDEX IF EXISTS idx_study_report_ref_skill_types_report_id;
DROP INDEX IF EXISTS idx_study_report_ref_skill_types_skill_type_id;
DROP INDEX IF EXISTS idx_study_report_ref_skill_types_skill_id;
DROP TABLE IF EXISTS study_report_ref_skill_types;

DROP INDEX IF EXISTS idx_study_report_skill_types_skill_id;
DROP TABLE IF EXISTS study_report_skill_types;

DROP INDEX IF EXISTS idx_study_report_skills_criteria_id;
DROP TABLE IF EXISTS study_report_skills;

DROP INDEX IF EXISTS idx_study_report_criterias_subject_id;
DROP TABLE IF EXISTS study_report_criterias;

DROP INDEX IF EXISTS idx_study_reports_subject_id;
DROP INDEX IF EXISTS idx_study_reports_criteria_id;
DROP TABLE IF EXISTS study_reports;
