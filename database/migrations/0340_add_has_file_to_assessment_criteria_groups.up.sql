ALTER TABLE assessment_criteria_groups
    ADD COLUMN IF NOT EXISTS has_file BOOLEAN DEFAULT FALSE;

