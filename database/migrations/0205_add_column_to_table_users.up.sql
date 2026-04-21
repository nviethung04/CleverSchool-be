ALTER TABLE users ADD COLUMN code TEXT;
CREATE INDEX idx_users_code ON users(code);
