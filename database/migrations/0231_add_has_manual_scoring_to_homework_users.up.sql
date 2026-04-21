ALTER TABLE homework_users
ADD COLUMN IF NOT EXISTS has_manual_scoring BOOLEAN DEFAULT false;


