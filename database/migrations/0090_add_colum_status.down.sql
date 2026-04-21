ALTER TABLE source_questions
    DROP COLUMN IF EXISTS status;

ALTER TABLE departments
    DROP COLUMN IF EXISTS status;

ALTER TABLE degrees
    DROP COLUMN IF EXISTS status;

ALTER TABLE certificates
    DROP COLUMN IF EXISTS status;

ALTER TABLE employee_positions
    DROP COLUMN IF EXISTS status;

ALTER TABLE roles
    DROP COLUMN IF EXISTS status;

ALTER TABLE schools
    DROP COLUMN IF EXISTS status;

ALTER TABLE question_attributes
    DROP COLUMN IF EXISTS status;
