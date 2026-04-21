CREATE TABLE study_reports (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    subject_id BIGINT NOT NULL,
    study_report_criteria_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    teacher_id BIGINT NOT NULL,
    general_comment TEXT,
    total_star SMALLINT,
    avg_star DECIMAL(5,2),
    type VARCHAR(50) NOT NULL CHECK (type IN ('month', 'semester')),
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    type_value TEXT,

    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

CREATE INDEX idx_study_reports_subject_id ON study_reports(subject_id);
CREATE INDEX idx_study_reports_criteria_id ON study_reports(study_report_criteria_id);


CREATE TABLE study_report_criterias (
    id SERIAL PRIMARY KEY,
    subject_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    max_star INT,
    notes JSONB DEFAULT '{}'::jsonb,

    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

CREATE INDEX idx_study_report_criterias_subject_id ON study_report_criterias(subject_id);


CREATE TABLE study_report_skills (
    id SERIAL PRIMARY KEY,
    criteria_id BIGINT NOT NULL,
    name_vn VARCHAR(255) NOT NULL,
    name_en VARCHAR(255),
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('star', 'check')),
    sort_order SMALLINT
);

CREATE INDEX idx_study_report_skills_criteria_id ON study_report_skills(criteria_id);


CREATE TABLE study_report_skill_types (
    id SERIAL PRIMARY KEY,
    skill_id BIGINT NOT NULL,
    name_vn VARCHAR(255) NOT NULL,
    name_en VARCHAR(255),
    sort_order SMALLINT
);

CREATE INDEX idx_study_report_skill_types_skill_id ON study_report_skill_types(skill_id);


CREATE TABLE study_report_ref_skill_types (
    id SERIAL PRIMARY KEY,
    study_report_id BIGINT NOT NULL,
    study_report_criteria_id BIGINT NOT NULL,
    skill_id BIGINT NOT NULL,
    skill_type_id BIGINT NOT NULL,
    note TEXT,
    star SMALLINT,
    is_check BOOLEAN DEFAULT FALSE
);

CREATE UNIQUE INDEX uq_study_report_ref_skill_types_report_skill
    ON study_report_ref_skill_types(study_report_id, skill_type_id);

CREATE INDEX idx_study_report_ref_skill_types_report_id
    ON study_report_ref_skill_types(study_report_id);

CREATE INDEX idx_study_report_ref_skill_types_skill_type_id
    ON study_report_ref_skill_types(skill_type_id);

CREATE INDEX idx_study_report_ref_skill_types_skill_id
    ON study_report_ref_skill_types(skill_id);

