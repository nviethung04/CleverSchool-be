DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT tablename
    FROM pg_tables
    WHERE schemaname = 'public'
      AND tablename ~ '^activity_logs_[0-9]{2}_[0-9]{4}$'
  LOOP
    EXECUTE format(
      'CREATE INDEX IF NOT EXISTS idx_%I_role_id_created_at ON %I (role_id, created_at)',
      r.tablename, r.tablename
    );
  END LOOP;
END $$;
