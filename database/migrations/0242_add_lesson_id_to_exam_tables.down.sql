-- Migration: Remove lesson_id column from all exam-related tables
-- Date: 2025-01-08

-- Remove lesson_id from exam_users
ALTER TABLE exam_users DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exam_question_users
ALTER TABLE exam_question_users DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exam_question_user_positions
ALTER TABLE exam_question_user_positions DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exam_question_user_matchings
ALTER TABLE exam_question_user_matchings DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exam_question_user_manual_scoring
ALTER TABLE exam_question_user_manual_scoring DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exam_question_user_labelings
ALTER TABLE exam_question_user_labelings DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exam_question_user_groups
ALTER TABLE exam_question_user_groups DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exam_question_user_fill_in_blanks
ALTER TABLE exam_question_user_fill_in_blanks DROP COLUMN IF EXISTS lesson_id;

-- Remove lesson_id from exam_comments
ALTER TABLE exam_comments DROP COLUMN IF EXISTS lesson_id;
