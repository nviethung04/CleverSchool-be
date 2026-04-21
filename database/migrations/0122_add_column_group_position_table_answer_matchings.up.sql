ALTER TABLE answer_matchings
ADD COLUMN group_position INTEGER DEFAULT 0;

UPDATE answer_matchings
SET group_position = correct_position;
