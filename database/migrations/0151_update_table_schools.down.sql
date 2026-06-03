DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'schools' AND column_name = 'ward_code'
    ) THEN
        ALTER TABLE schools RENAME COLUMN ward_code TO district_code;
    END IF;
END $$;
