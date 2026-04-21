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
