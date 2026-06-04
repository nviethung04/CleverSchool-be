CREATE EXTENSION IF NOT EXISTS unaccent;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'kind_enum') THEN
        CREATE TYPE kind_enum AS ENUM ('text', 'audio', 'image', 'video');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'level_enum') THEN
        CREATE TYPE level_enum AS ENUM ('easy', 'medium', 'hard');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'question_type_enum') THEN
        CREATE TYPE question_type_enum AS ENUM (
            'multiple_choice', 'fill_in_blanks', 'ordering', 'matching',
            'drag_drop', 'labeling', 'category', 'speaking', 'writing'
        );
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'display_enum') THEN
        CREATE TYPE display_enum AS ENUM ('word_count', 'vertical', 'horizontal');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'course_state') THEN
        CREATE TYPE course_state AS ENUM ('coming', 'active', 'finished');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'question_form_enum') THEN
        CREATE TYPE question_form_enum AS ENUM ('write', 'question');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'join_level_enum') THEN
        CREATE TYPE join_level_enum AS ENUM ('school', 'province', 'person', 'class');
    END IF;
END $$;
