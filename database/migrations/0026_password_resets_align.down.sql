DROP INDEX IF EXISTS idx_password_resets_expires_at;
DROP INDEX IF EXISTS idx_password_resets_status;
DROP INDEX IF EXISTS idx_password_resets_token;
DROP INDEX IF EXISTS idx_password_resets_email;
DROP INDEX IF EXISTS idx_password_resets_user_id;

ALTER TABLE password_resets DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE password_resets DROP COLUMN IF EXISTS updated_at;
ALTER TABLE password_resets DROP COLUMN IF EXISTS used_at;
ALTER TABLE password_resets DROP COLUMN IF EXISTS expires_at;
ALTER TABLE password_resets DROP COLUMN IF EXISTS status;
ALTER TABLE password_resets DROP COLUMN IF EXISTS user_id;
