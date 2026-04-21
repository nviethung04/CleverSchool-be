ALTER TABLE roles DROP COLUMN IF EXISTS default_page_view;
ALTER TABLE roles DROP COLUMN IF EXISTS default_page_id;

UPDATE roles
SET name = 'Read only'
WHERE name IN ('read only');
