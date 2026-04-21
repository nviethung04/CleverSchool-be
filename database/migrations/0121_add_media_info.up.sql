ALTER TABLE questions
ADD COLUMN media_info JSONB;

ALTER TABLE answer_groups
ADD COLUMN media_info JSONB;

ALTER TABLE answer_coordinates
ADD COLUMN media_info JSONB;

ALTER TABLE answer_matchings
ADD COLUMN media_info JSONB;

ALTER TABLE answer_matchings
ADD COLUMN matching_media_info JSONB;

ALTER TABLE answer_positions
ADD COLUMN media_info JSONB;

ALTER TABLE answers
ADD COLUMN media_info JSONB;

ALTER TABLE group_answers
ADD COLUMN media_info JSONB;
