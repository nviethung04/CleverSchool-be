ALTER TABLE users
ADD COLUMN avatar_info JSONB;

ALTER TABLE certificates
ADD COLUMN file_info JSONB;

ALTER TABLE courses
ADD COLUMN image_info JSONB;

ALTER TABLE degrees
ADD COLUMN file_info JSONB;

ALTER TABLE exam_question_user_manual_scoring
ADD COLUMN file_info JSONB;

ALTER TABLE exams
ADD COLUMN cover_image_info JSONB;

ALTER TABLE homework_question_user_manual_scoring
ADD COLUMN file_info JSONB;

ALTER TABLE homeworks
ADD COLUMN cover_image_info JSONB;

ALTER TABLE lesson_plan_parts
ADD COLUMN link_info JSONB;

ALTER TABLE lesson_plans
ADD COLUMN cover_image_info JSONB;

ALTER TABLE level_tests
ADD COLUMN cover_image_info JSONB;

ALTER TABLE schools
ADD COLUMN logo_info JSONB;

ALTER TABLE skills
ADD COLUMN image_info JSONB;

ALTER TABLE tags
ADD COLUMN image_info JSONB;

ALTER TABLE topics
ADD COLUMN image_info JSONB;

ALTER TABLE questions
ADD COLUMN file_info JSONB;

ALTER TABLE answer_groups
ADD COLUMN file_info JSONB;

ALTER TABLE answer_coordinates
ADD COLUMN file_info JSONB;

ALTER TABLE answer_matchings
ADD COLUMN file_info JSONB;

ALTER TABLE answer_positions
ADD COLUMN file_info JSONB;

ALTER TABLE answers
ADD COLUMN file_info JSONB;

ALTER TABLE group_answers
ADD COLUMN file_info JSONB;

ALTER TABLE answer_matchings
ADD COLUMN matching_file_info JSONB;

ALTER TABLE questions
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_groups
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_coordinates
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_matchings
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_positions
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answers
DROP COLUMN IF EXISTS media_info;

ALTER TABLE group_answers
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_matchings
DROP COLUMN IF EXISTS matching_media_info;
