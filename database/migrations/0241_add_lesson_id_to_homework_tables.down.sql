-- Migration: Remove lesson_id column from all homework-related tables
-- Date: 2025-01-08

-- Remove lesson_id from homework_comments
ALTER TABLE IF EXISTS homework_comments DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from homework_question_user_fill_in_blanks
ALTER TABLE IF EXISTS homework_question_user_fill_in_blanks DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from homework_question_user_groups
ALTER TABLE IF EXISTS homework_question_user_groups DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from homework_question_user_labelings
ALTER TABLE IF EXISTS homework_question_user_labelings DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from homework_question_user_manual_scoring
ALTER TABLE IF EXISTS homework_question_user_manual_scoring DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from homework_question_user_matchings
ALTER TABLE IF EXISTS homework_question_user_matchings DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from homework_question_user_positions
ALTER TABLE IF EXISTS homework_question_user_positions DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from homework_question_users
ALTER TABLE IF EXISTS homework_question_users DROP COLUMN IF EXISTS lesson_id;

-- Skip homework_ref_lessons - already has lesson_id column

-- Remove lesson_id from homework_user_skip_questions
ALTER TABLE IF EXISTS homework_user_skip_questions DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from homework_users
ALTER TABLE IF EXISTS homework_users DROP COLUMN IF EXISTS lesson_id;

