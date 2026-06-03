DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'level_tests' AND column_name = 'level_standard_id'
    ) THEN
        ALTER TABLE level_tests RENAME COLUMN level_standard_id TO level_type_id;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'levels' AND column_name = 'level_standard_id'
    ) THEN
        ALTER TABLE levels RENAME COLUMN level_standard_id TO level_type_id;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'level_standards'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'level_types'
    ) THEN
        ALTER TABLE level_standards RENAME TO level_types;
    END IF;
END $$;
