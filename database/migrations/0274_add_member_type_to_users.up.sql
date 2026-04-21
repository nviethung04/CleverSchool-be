ALTER TABLE users
ADD COLUMN member_type VARCHAR(20) DEFAULT 'internal' NOT NULL;

-- Add check constraint to ensure member_type is either 'external' or 'internal'
ALTER TABLE users
ADD CONSTRAINT check_member_type CHECK (member_type IN ('external', 'internal'));

-- Create index for member_type
CREATE INDEX idx_users_member_type ON users(member_type);

