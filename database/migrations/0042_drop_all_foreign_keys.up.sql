DO $$
DECLARE
r RECORD;
BEGIN
FOR r IN
SELECT conrelid::regclass AS table_name,
        conname
FROM pg_constraint
WHERE contype = 'f'
    LOOP
        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT %I;', r.table_name, r.conname);
END LOOP;
END $$;
