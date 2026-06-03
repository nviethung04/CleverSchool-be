DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'h5p_contents'
          AND column_name = 'id'
          AND is_identity = 'NO'
    ) THEN
        ALTER TABLE h5p_contents ALTER COLUMN id SET DEFAULT nextval('h5p_contents_id_seq');
    END IF;
END $$;
