-- Align password_resets with models.PasswordReset (baseline 0013 was minimal).

ALTER TABLE password_resets ADD COLUMN IF NOT EXISTS user_id BIGINT;
ALTER TABLE password_resets ADD COLUMN IF NOT EXISTS status BOOLEAN DEFAULT TRUE;
ALTER TABLE password_resets ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP;
ALTER TABLE password_resets ADD COLUMN IF NOT EXISTS used_at TIMESTAMP;
ALTER TABLE password_resets ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE password_resets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

UPDATE password_resets SET status = TRUE WHERE status IS NULL;

UPDATE password_resets
SET expires_at = created_at + INTERVAL '24 hours'
WHERE expires_at IS NULL AND created_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets (user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_email ON password_resets (email);
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets (token);
CREATE INDEX IF NOT EXISTS idx_password_resets_status ON password_resets (status);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets (expires_at);
