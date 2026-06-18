CREATE TABLE assessments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL DEFAULT 'mini_test',
    program_id BIGINT REFERENCES programs(id) ON DELETE SET NULL,
    subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    assessment_criteria_group_id BIGINT,
    study_report_criteria_id BIGINT,
    assessment_criteria_ids JSONB NOT NULL DEFAULT '[]',
    file_infos JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE INDEX idx_assessments_subject_id ON assessments(subject_id);
CREATE INDEX idx_assessments_deleted_at ON assessments(deleted_at);
