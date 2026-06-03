-- Migration: Remove lesson_id column from all exercise-related tables
-- Date: 2025-01-08

-- Remove lesson_id from exercise_users
ALTER TABLE exercise_users DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_questions
ALTER TABLE exercise_questions DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_question_users
ALTER TABLE exercise_question_users DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_question_user_positions
ALTER TABLE exercise_question_user_positions DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_question_user_matchings
ALTER TABLE exercise_question_user_matchings DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_question_user_manual_scoring
ALTER TABLE exercise_question_user_manual_scoring DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_question_user_labelings
ALTER TABLE exercise_question_user_labelings DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_question_user_groups
ALTER TABLE exercise_question_user_groups DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_question_user_fill_in_blanks
ALTER TABLE exercise_question_user_fill_in_blanks DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exercise_comments
ALTER TABLE exercise_comments DROP COLUMN IF EXISTS lesson_id;
