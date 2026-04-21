ALTER TABLE roles
DROP COLUMN IF EXISTS parent_id;

DELETE FROM roles
WHERE name = 'school';
