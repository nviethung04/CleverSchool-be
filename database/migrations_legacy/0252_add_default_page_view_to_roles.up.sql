ALTER TABLE roles ADD COLUMN IF NOT EXISTS default_page_view VARCHAR(25);
ALTER TABLE roles ADD COLUMN IF NOT EXISTS default_page_id INT;

UPDATE roles
SET
  default_page_view = 'admin',
  default_page_id = 1
WHERE name IN ('admin', 'school', 'Read only');

UPDATE roles
SET
    default_page_view = 'teacher',
    default_page_id = 2
WHERE name IN ('teacher');

UPDATE roles
SET
    default_page_view = 'student',
    default_page_id = 3
WHERE name IN ('student');

UPDATE roles
SET name = 'read only'
WHERE name IN ('Read only');
