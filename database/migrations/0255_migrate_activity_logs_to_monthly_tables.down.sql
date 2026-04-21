-- Rollback migration: Migrate activity_logs to monthly tables
-- This migration drops monthly tables and restores the original structure

-- 1. Drop all monthly activity_logs tables
DO $$
DECLARE
    rec RECORD;
BEGIN
    -- Lấy danh sách tất cả bảng monthly activity_logs
    FOR rec IN 
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name LIKE 'activity_logs_%_%'
        ORDER BY table_name
    LOOP
        EXECUTE format('DROP TABLE IF EXISTS %I', rec.table_name);
        RAISE NOTICE 'Dropped table: %', rec.table_name;
    END LOOP;
    
    RAISE NOTICE 'All monthly activity_logs tables dropped';
END $$;

-- 2. Drop helper functions
DROP FUNCTION IF EXISTS create_monthly_activity_logs_table(DATE);
DROP FUNCTION IF EXISTS cleanup_old_activity_logs_table();

-- 3. Recreate original activity_logs table if it doesn't exist
CREATE TABLE IF NOT EXISTS activity_logs (
    id SERIAL PRIMARY KEY,
    user_id BIGINT,
    role_id INTEGER,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    client_ip VARCHAR(45),
    device VARCHAR(20),
    latency_ms INTEGER,
    query_params JSONB,
    request_body JSONB,
    response_body JSONB,
    attribute JSONB,
    session_id TEXT,
    agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 4. Recreate original indexes
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_created_at ON activity_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_device ON activity_logs(device);
CREATE INDEX IF NOT EXISTS idx_activity_logs_device_created_at ON activity_logs(device, created_at DESC);
