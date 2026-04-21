ALTER TABLE programs
ADD COLUMN type VARCHAR(20) DEFAULT 'theory' NOT NULL;

-- Add check constraint to ensure type is either 'theory', 'practice' or 'module'
ALTER TABLE programs
ADD CONSTRAINT check_program_type CHECK (type IN ('theory', 'practice', 'module'));

-- Create index for type
CREATE INDEX idx_programs_type ON programs(type);
