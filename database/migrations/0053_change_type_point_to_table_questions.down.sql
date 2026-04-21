ALTER TABLE questions
    ALTER COLUMN point TYPE integer USING ROUND(point);

ALTER TABLE answer_coordinates
    ALTER COLUMN point TYPE integer USING ROUND(point);

ALTER TABLE answer_groups
    ALTER COLUMN point TYPE integer USING ROUND(point);

ALTER TABLE answer_matchings
    ALTER COLUMN point TYPE integer USING ROUND(point);

ALTER TABLE answer_positions
    ALTER COLUMN point TYPE integer USING ROUND(point);

ALTER TABLE answers
    ALTER COLUMN point TYPE integer USING ROUND(point);
