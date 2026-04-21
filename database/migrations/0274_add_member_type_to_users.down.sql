DROP INDEX IF EXISTS idx_users_member_type;
ALTER TABLE users
DROP CONSTRAINT IF EXISTS check_member_type;
ALTER TABLE users
DROP COLUMN IF EXISTS member_type;

