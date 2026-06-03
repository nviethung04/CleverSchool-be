-- Migration: Add lesson_id column to all homework-related tables
-- Date: 2025-01-08

-- Add lesson_id to homework_users (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_users' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_users ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to homework_user_skip_questions (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_user_skip_questions' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_user_skip_questions ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Skip homework_ref_lessons - already has lesson_id column

-- Add lesson_id to homework_question_users (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_question_users' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_question_users ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to homework_question_user_positions (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_question_user_positions' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_question_user_positions ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to homework_question_user_matchings (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_question_user_matchings' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_question_user_matchings ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to homework_question_user_manual_scoring (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_question_user_manual_scoring' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_question_user_manual_scoring ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to homework_question_user_labelings (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_question_user_labelings' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_question_user_labelings ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to homework_question_user_groups (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_question_user_groups' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_question_user_groups ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to homework_question_user_fill_in_blanks (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_question_user_fill_in_blanks' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_question_user_fill_in_blanks ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to homework_comments (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'homework_comments' AND column_name = 'lesson_id') THEN
        ALTER TABLE homework_comments ADD COLUMN lesson_id bigint;
    END IF;
END $$;
