DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'schools' AND column_name = 'contact_mail'
    ) THEN
        ALTER TABLE schools
        ALTER COLUMN contact_mail TYPE VARCHAR(50);
    END IF;
END $$;
