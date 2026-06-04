ALTER TABLE answer_matchings
ADD COLUMN matching_group_position INTEGER DEFAULT 0;

UPDATE answer_matchings
SET matching_group_position = correct_position;
