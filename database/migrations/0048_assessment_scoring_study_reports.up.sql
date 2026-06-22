-- Assessment criteria (chấm điểm theo tiêu chí)
CREATE TABLE IF NOT EXISTS assessment_criteria_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    has_file BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE INDEX IF NOT EXISTS idx_assessment_criteria_groups_subject_id ON assessment_criteria_groups(subject_id);
CREATE INDEX IF NOT EXISTS idx_assessment_criteria_groups_deleted_at ON assessment_criteria_groups(deleted_at);

CREATE TABLE IF NOT EXISTS assessment_criteria (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES assessment_criteria_groups(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    has_subcriteria BOOLEAN NOT NULL DEFAULT FALSE,
    max_score NUMERIC(10, 2) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE INDEX IF NOT EXISTS idx_assessment_criteria_group_id ON assessment_criteria(group_id);

CREATE TABLE IF NOT EXISTS assessment_subcriteria (
    id BIGSERIAL PRIMARY KEY,
    criteria_id BIGINT NOT NULL REFERENCES assessment_criteria(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    max_score NUMERIC(10, 2) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE INDEX IF NOT EXISTS idx_assessment_subcriteria_criteria_id ON assessment_subcriteria(criteria_id);

-- Tiêu chí báo cáo học tập (nhận xét sao / checklist)
CREATE TABLE IF NOT EXISTS study_report_criterias (
    id BIGSERIAL PRIMARY KEY,
    subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    max_star INT NOT NULL DEFAULT 5,
    notes JSONB NOT NULL DEFAULT '[]',
    star_skills JSONB NOT NULL DEFAULT '[]',
    check_skills JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_study_report_criterias_subject_id ON study_report_criterias(subject_id);

-- Điểm assessment theo học sinh / khóa
CREATE TABLE IF NOT EXISTS assessment_scores (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assessment_id BIGINT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    total_score NUMERIC(10, 2) NOT NULL DEFAULT 0,
    file_infos JSONB,
    is_scored BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    UNIQUE (user_id, assessment_id, course_id)
);

CREATE INDEX IF NOT EXISTS idx_assessment_scores_assessment_course ON assessment_scores(assessment_id, course_id);

CREATE TABLE IF NOT EXISTS assessment_score_details (
    id BIGSERIAL PRIMARY KEY,
    assessment_score_id BIGINT NOT NULL REFERENCES assessment_scores(id) ON DELETE CASCADE,
    criteria_id BIGINT NOT NULL REFERENCES assessment_criteria(id) ON DELETE CASCADE,
    subcriteria_id BIGINT REFERENCES assessment_subcriteria(id) ON DELETE CASCADE,
    score NUMERIC(10, 2) NOT NULL DEFAULT -1
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_assessment_score_details_unique
    ON assessment_score_details(assessment_score_id, criteria_id, COALESCE(subcriteria_id, 0));

CREATE TABLE IF NOT EXISTS assessment_publishes (
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    assessment_id BIGINT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    publish BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (course_id, assessment_id)
);

-- Báo cáo học tập (nhận xét GV)
CREATE TABLE IF NOT EXISTS study_reports (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    study_report_criteria_id BIGINT REFERENCES study_report_criterias(id) ON DELETE SET NULL,
    assessment_id BIGINT REFERENCES assessments(id) ON DELETE SET NULL,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    teacher_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    general_comment TEXT,
    total_star NUMERIC(10, 2) DEFAULT 0,
    avg_star NUMERIC(10, 2) DEFAULT 0,
    type VARCHAR(50),
    type_value VARCHAR(255),
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    max_star INT NOT NULL DEFAULT 5,
    star_skills JSONB NOT NULL DEFAULT '[]',
    check_skills JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (assessment_id, course_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_study_reports_course_assessment ON study_reports(course_id, assessment_id);

CREATE TABLE IF NOT EXISTS study_report_publishes (
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    assessment_id BIGINT NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    publish BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (course_id, assessment_id)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_assessments_criteria_group'
    ) THEN
        ALTER TABLE assessments
            ADD CONSTRAINT fk_assessments_criteria_group
            FOREIGN KEY (assessment_criteria_group_id)
            REFERENCES assessment_criteria_groups(id) ON DELETE SET NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_assessments_study_report_criteria'
    ) THEN
        ALTER TABLE assessments
            ADD CONSTRAINT fk_assessments_study_report_criteria
            FOREIGN KEY (study_report_criteria_id)
            REFERENCES study_report_criterias(id) ON DELETE SET NULL;
    END IF;
END $$;
