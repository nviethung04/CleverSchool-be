-- 1. source_questions
ALTER TABLE source_questions
DROP COLUMN IF EXISTS status;

ALTER TABLE source_questions
ADD COLUMN status BOOLEAN NOT NULL DEFAULT true;

-- 2. departments
ALTER TABLE departments
DROP COLUMN IF EXISTS status;

ALTER TABLE departments
ADD COLUMN status BOOLEAN NOT NULL DEFAULT true;

-- 3. degrees
ALTER TABLE degrees
DROP COLUMN IF EXISTS status;

ALTER TABLE degrees
ADD COLUMN status BOOLEAN NOT NULL DEFAULT true;

-- 4. certificates
ALTER TABLE certificates
DROP COLUMN IF EXISTS status;

ALTER TABLE certificates
ADD COLUMN status BOOLEAN NOT NULL DEFAULT true;

-- 5. employee_positions
ALTER TABLE employee_positions
DROP COLUMN IF EXISTS status;

ALTER TABLE employee_positions
ADD COLUMN status BOOLEAN NOT NULL DEFAULT true;

-- 6. roles
ALTER TABLE roles
DROP COLUMN IF EXISTS status;

ALTER TABLE roles
ADD COLUMN status BOOLEAN NOT NULL DEFAULT true;

-- 7. schools
ALTER TABLE schools
DROP COLUMN IF EXISTS status;

ALTER TABLE schools
ADD COLUMN status BOOLEAN NOT NULL DEFAULT true;

-- 8. question_attributes
ALTER TABLE question_attributes
DROP COLUMN IF EXISTS status;

ALTER TABLE question_attributes
ADD COLUMN status BOOLEAN NOT NULL DEFAULT true;
