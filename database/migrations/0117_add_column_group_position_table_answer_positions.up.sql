ALTER TABLE answer_positions
ADD COLUMN group_position INTEGER DEFAULT 0;

UPDATE answer_positions
SET group_position = correct_position;

