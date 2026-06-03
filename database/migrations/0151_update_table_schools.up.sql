DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'schools' AND column_name = 'district_code'
    ) THEN
        ALTER TABLE schools RENAME COLUMN district_code TO ward_code;
    END IF;
END $$;
