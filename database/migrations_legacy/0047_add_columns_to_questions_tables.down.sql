-- Drop columns from homework_questions
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'homework_questions' AND column_name = 'score'
    ) THEN
        ALTER TABLE homework_questions DROP COLUMN score;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'homework_questions' AND column_name = 'is_source_question'
    ) THEN
        ALTER TABLE homework_questions DROP COLUMN is_source_question;
    END IF;
END$$;

-- Drop columns from exam_questions
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'exam_questions' AND column_name = 'score'
    ) THEN
        ALTER TABLE exam_questions DROP COLUMN score;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'exam_questions' AND column_name = 'is_source_question'
    ) THEN
        ALTER TABLE exam_questions DROP COLUMN is_source_question;
    END IF;
END$$;

-- Drop columns from lesson_plan_part_questions
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'lesson_plan_part_questions' AND column_name = 'score'
    ) THEN
        ALTER TABLE lesson_plan_part_questions DROP COLUMN score;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'lesson_plan_part_questions' AND column_name = 'is_source_question'
    ) THEN
        ALTER TABLE lesson_plan_part_questions DROP COLUMN is_source_question;
    END IF;
END$$;
