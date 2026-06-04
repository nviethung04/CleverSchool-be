-- Alter score column type in exams table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'exams' AND column_name = 'score'
    ) THEN
        ALTER TABLE public.exams
        ALTER COLUMN score TYPE numeric(5,2) USING score::numeric(5,2);
    END IF;
END$$;

-- Alter score column type in homeworks table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'homeworks' AND column_name = 'score'
    ) THEN
        ALTER TABLE public.homeworks
        ALTER COLUMN score TYPE numeric(5,2) USING score::numeric(5,2);
    END IF;
END$$;
