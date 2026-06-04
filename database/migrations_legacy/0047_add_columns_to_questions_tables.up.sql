-- Add columns to homework_questions
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'homework_questions' AND column_name = 'score'
    ) THEN
        ALTER TABLE homework_questions ADD COLUMN score numeric(5,2);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'homework_questions' AND column_name = 'is_source_question'
    ) THEN
        ALTER TABLE homework_questions ADD COLUMN is_source_question boolean DEFAULT false;
    END IF;
END$$;

-- Add columns to exam_questions
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'exam_questions' AND column_name = 'score'
    ) THEN
        ALTER TABLE exam_questions ADD COLUMN score numeric(5,2);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'exam_questions' AND column_name = 'is_source_question'
    ) THEN
        ALTER TABLE exam_questions ADD COLUMN is_source_question boolean DEFAULT false;
    END IF;
END$$;

-- Add columns to lesson_plan_part_questions
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'lesson_plan_part_questions' AND column_name = 'score'
    ) THEN
        ALTER TABLE lesson_plan_part_questions ADD COLUMN score numeric(5,2);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'lesson_plan_part_questions' AND column_name = 'is_source_question'
    ) THEN
        ALTER TABLE lesson_plan_part_questions ADD COLUMN is_source_question boolean DEFAULT false;
    END IF;
END$$;
