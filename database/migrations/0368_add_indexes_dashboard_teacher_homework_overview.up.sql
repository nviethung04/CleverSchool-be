DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_users')
       AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='homework_users' AND column_name='updated_at') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_users_updated_at ON homework_users (updated_at);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='deleted_at') THEN
        CREATE INDEX IF NOT EXISTS idx_users_deleted_at_null ON users (id) WHERE deleted_at IS NULL;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='homeworks' AND column_name='deleted_at') THEN
        CREATE INDEX IF NOT EXISTS idx_homeworks_deleted_at_null ON homeworks (id) WHERE deleted_at IS NULL;
    END IF;
END $$;
