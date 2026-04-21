ALTER TABLE programs
DROP CONSTRAINT IF EXISTS check_program_type;

ALTER TABLE programs
ADD CONSTRAINT check_program_type
CHECK (type IN ('theory', 'practice', 'module'));
