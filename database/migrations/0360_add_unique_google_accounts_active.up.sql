-- Ensure only one active Google account per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_google_accounts_user_active
    ON google_accounts(user_id)
    WHERE is_active;
