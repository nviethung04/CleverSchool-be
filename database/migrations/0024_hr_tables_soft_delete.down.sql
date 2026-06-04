ALTER TABLE employee_positions DROP COLUMN IF EXISTS deleted_by, DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE departments DROP COLUMN IF EXISTS deleted_by, DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE degrees DROP COLUMN IF EXISTS deleted_by, DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE certificates DROP COLUMN IF EXISTS deleted_by, DROP COLUMN IF EXISTS deleted_at;
