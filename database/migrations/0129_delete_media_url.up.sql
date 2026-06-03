ALTER TABLE users
DROP COLUMN IF EXISTS avatar;

ALTER TABLE certificates
DROP COLUMN IF EXISTS file_url;

ALTER TABLE courses
DROP COLUMN IF EXISTS image;

ALTER TABLE degrees
DROP COLUMN IF EXISTS file_url;

ALTER TABLE exam_question_user_manual_scoring
DROP COLUMN IF EXISTS file_url;

ALTER TABLE exams
DROP COLUMN IF EXISTS cover_image;

ALTER TABLE homework_question_user_manual_scoring
DROP COLUMN IF EXISTS file_url;

ALTER TABLE homeworks
DROP COLUMN IF EXISTS cover_image;

ALTER TABLE lesson_plan_parts
DROP COLUMN IF EXISTS link;

ALTER TABLE lesson_plans
DROP COLUMN IF EXISTS cover_image;

ALTER TABLE level_tests
DROP COLUMN IF EXISTS cover_image;

ALTER TABLE schools
DROP COLUMN IF EXISTS logo;

ALTER TABLE skills
DROP COLUMN IF EXISTS image_url;

ALTER TABLE tags
DROP COLUMN IF EXISTS image_url;

ALTER TABLE topics
DROP COLUMN IF EXISTS image_url;

ALTER TABLE questions
DROP COLUMN IF EXISTS file_url;

ALTER TABLE answer_groups
DROP COLUMN IF EXISTS file_url;

ALTER TABLE answer_coordinates
DROP COLUMN IF EXISTS media_url;

ALTER TABLE answer_matchings
DROP COLUMN IF EXISTS file_url;

ALTER TABLE answer_positions
DROP COLUMN IF EXISTS media_url;

ALTER TABLE answers
DROP COLUMN IF EXISTS file_url;

ALTER TABLE group_answers
DROP COLUMN IF EXISTS file_url;

ALTER TABLE answer_matchings
DROP COLUMN IF EXISTS matching_file_url;

ALTER TABLE lesson_plan_parts
DROP COLUMN IF EXISTS cover_image;
