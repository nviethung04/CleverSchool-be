-- +migrate Up

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'kind_enum') THEN
CREATE TYPE kind_enum AS ENUM ('text', 'audio', 'image', 'video');
END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'level_enum') THEN
CREATE TYPE level_enum AS ENUM ('easy', 'medium', 'hard');
END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'question_type_enum') THEN
CREATE TYPE question_type_enum AS ENUM ('multiple_choice', 'fill_in_blanks', 'ordering', 'matching', 'drag_drop', 'labeling', 'category', 'speaking', 'writing');
END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'skill_enum') THEN
CREATE TYPE skill_enum AS ENUM ('Listening', 'Reading', 'Writing', 'Speaking', 'Use of English');
END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'topic_enum') THEN
CREATE TYPE topic_enum AS ENUM ('grammar', 'vocabulary');
END IF;
END $$;