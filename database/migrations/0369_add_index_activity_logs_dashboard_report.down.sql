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
    EXECUTE format('DROP INDEX IF EXISTS idx_%I_role_id_created_at', r.tablename);
  END LOOP;
END $$;
