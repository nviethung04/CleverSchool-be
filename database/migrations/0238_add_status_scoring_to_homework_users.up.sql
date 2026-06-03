DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_users') THEN
        ALTER TABLE IF EXISTS homework_users
        ADD COLUMN IF NOT EXISTS status_scoring SMALLINT DEFAULT 0;

        COMMENT ON COLUMN homework_users.status_scoring IS '0: khong can cham, 1: chua cham xong, 2: da cham xong';

        CREATE INDEX IF NOT EXISTS idx_homework_users_status_scoring
        ON homework_users(status_scoring);
    END IF;
END $$;
