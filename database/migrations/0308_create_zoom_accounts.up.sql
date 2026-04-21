-- Create zoom_accounts table
CREATE TABLE IF NOT EXISTS zoom_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    zoom_user_id VARCHAR(255),
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP,
    account_email VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    
    CONSTRAINT fk_zoom_accounts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_zoom_accounts_user_id ON zoom_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_zoom_accounts_is_active ON zoom_accounts(is_active);
CREATE INDEX IF NOT EXISTS idx_zoom_accounts_deleted_at ON zoom_accounts(deleted_at);

-- Create unique constraint for active accounts per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_zoom_accounts_user_active 
ON zoom_accounts(user_id) WHERE is_active = true AND deleted_at IS NULL;