ALTER TABLE source_questions
    ADD COLUMN file_infos JSONB;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'default' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'skill_enum')) THEN
        ALTER TYPE skill_enum ADD VALUE 'default';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'default' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'level_enum')) THEN
        ALTER TYPE level_enum ADD VALUE 'default';
    END IF;
END $$;
