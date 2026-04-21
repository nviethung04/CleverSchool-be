DROP INDEX IF EXISTS idx_programs_type;
ALTER TABLE programs
DROP CONSTRAINT IF EXISTS check_program_type;
ALTER TABLE programs
DROP COLUMN IF EXISTS type;

