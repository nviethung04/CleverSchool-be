DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'history_uses_type_date_unique'
  ) THEN
    ALTER TABLE history_uses
      ADD CONSTRAINT history_uses_type_date_unique UNIQUE (type, date);
  END IF;
END;
$$;
