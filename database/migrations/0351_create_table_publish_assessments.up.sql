CREATE TABLE IF NOT EXISTS publish_assessments (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT,
    assessment_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_publish_assessments UNIQUE (course_id, assessment_id)
);

CREATE INDEX idx_publish_assessments_course_id ON publish_assessments(course_id);
CREATE INDEX idx_publish_assessments_assessment_id ON publish_assessments(assessment_id);

