CREATE TABLE IF NOT EXISTS publish_reports (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT,
    assessment_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_publish_reports UNIQUE (course_id, assessment_id)
);

CREATE INDEX idx_publish_reports_course_id ON publish_reports(course_id);
CREATE INDEX idx_publish_reports_assessment_id ON publish_reports(assessment_id);

