ALTER TABLE study_report_skill_types
    ADD COLUMN parent_id BIGINT,
    ADD COLUMN level SMALLINT NOT NULL DEFAULT 1;
