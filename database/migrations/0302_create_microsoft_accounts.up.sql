CREATE TABLE IF NOT EXISTS microsoft_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    microsoft_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    account_email VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_microsoft_accounts_user_id ON microsoft_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_microsoft_accounts_deleted_at ON microsoft_accounts(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_microsoft_accounts_user_microsoft ON microsoft_accounts(user_id, microsoft_user_id) WHERE deleted_at IS NULL;