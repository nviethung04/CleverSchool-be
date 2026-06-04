DROP INDEX IF EXISTS idx_certificates_course_id;

ALTER TABLE certificates DROP COLUMN IF EXISTS course_id;
ALTER TABLE certificates DROP COLUMN IF EXISTS grade;
ALTER TABLE certificates DROP COLUMN IF EXISTS rating;
