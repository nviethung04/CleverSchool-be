ALTER TABLE permissions
ADD COLUMN sort_position INT;

ALTER TABLE permissions
ALTER COLUMN sort_position SET DEFAULT 0;