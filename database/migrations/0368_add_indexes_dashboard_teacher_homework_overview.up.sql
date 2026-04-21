CREATE INDEX IF NOT EXISTS idx_homework_users_updated_at
    ON homework_users (updated_at);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at_null
    ON users (id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_homeworks_deleted_at_null
    ON homeworks (id) WHERE deleted_at IS NULL;
