ALTER TABLE certificates ADD COLUMN IF NOT EXISTS course_id BIGINT DEFAULT 0;
ALTER TABLE certificates ADD COLUMN IF NOT EXISTS grade VARCHAR(50);
ALTER TABLE certificates ADD COLUMN IF NOT EXISTS rating NUMERIC(3,2);

CREATE INDEX IF NOT EXISTS idx_certificates_course_id ON certificates(course_id);
