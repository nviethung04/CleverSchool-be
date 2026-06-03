-- Migration: Add lesson_id column to all exercise-related tables
-- Date: 2025-01-08

-- Add lesson_id to exercise_comments (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_comments' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_comments ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_question_user_fill_in_blanks (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_question_user_fill_in_blanks' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_question_user_fill_in_blanks ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_question_user_groups (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_question_user_groups' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_question_user_groups ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_question_user_labelings (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_question_user_labelings' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_question_user_labelings ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_question_user_manual_scoring (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_question_user_manual_scoring' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_question_user_manual_scoring ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_question_user_matchings (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_question_user_matchings' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_question_user_matchings ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_question_user_positions (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_question_user_positions' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_question_user_positions ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_question_users (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_question_users' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_question_users ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_questions (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_questions' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_questions ADD COLUMN lesson_id bigint;
    END IF;
END $$;

-- Add lesson_id to exercise_users (if not exists)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'exercise_users' AND column_name = 'lesson_id') THEN
        ALTER TABLE exercise_users ADD COLUMN lesson_id bigint;
    END IF;
END $$;
