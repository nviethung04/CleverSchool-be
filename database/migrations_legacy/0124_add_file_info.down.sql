ALTER TABLE users
DROP COLUMN IF EXISTS avatar_info;

ALTER TABLE certificates
DROP COLUMN IF EXISTS file_info;

ALTER TABLE courses
DROP COLUMN IF EXISTS image_info;

ALTER TABLE degrees
DROP COLUMN IF EXISTS file_info;

ALTER TABLE exam_question_user_manual_scoring
DROP COLUMN IF EXISTS file_info;

ALTER TABLE exams
DROP COLUMN IF EXISTS cover_image_info;

ALTER TABLE homework_question_user_manual_scoring
DROP COLUMN IF EXISTS file_info;

ALTER TABLE homeworks
DROP COLUMN IF EXISTS cover_image_info;

ALTER TABLE lesson_plan_parts
DROP COLUMN IF EXISTS link_info;

ALTER TABLE lesson_plans
DROP COLUMN IF EXISTS cover_image_info;

ALTER TABLE level_tests
DROP COLUMN IF EXISTS cover_image_info;

ALTER TABLE schools
DROP COLUMN IF EXISTS logo_info;

ALTER TABLE skills
DROP COLUMN IF EXISTS image_info;

ALTER TABLE tags
DROP COLUMN IF EXISTS image_info;

ALTER TABLE topics
DROP COLUMN IF EXISTS image_info;

ALTER TABLE questions
DROP COLUMN IF EXISTS file_info;

ALTER TABLE answer_groups
DROP COLUMN IF EXISTS file_info;

ALTER TABLE answer_coordinates
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_matchings
DROP COLUMN IF EXISTS file_info;

ALTER TABLE answer_positions
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answers
DROP COLUMN IF EXISTS file_info;

ALTER TABLE group_answers
DROP COLUMN IF EXISTS file_info;

ALTER TABLE answer_matchings
DROP COLUMN IF EXISTS matching_file_info;

ALTER TABLE questions
ADD COLUMN media_info JSONB;

ALTER TABLE answer_groups
ADD COLUMN media_info JSONB;

ALTER TABLE answer_coordinates
ADD COLUMN media_info JSONB;

ALTER TABLE answer_matchings
ADD COLUMN media_info JSONB;

ALTER TABLE answer_positions
ADD COLUMN media_info JSONB;

ALTER TABLE answers
ADD COLUMN media_info JSONB;

ALTER TABLE group_answers
ADD COLUMN media_info JSONB;

ALTER TABLE answer_matchings
ADD COLUMN matching_media_info JSONB;
