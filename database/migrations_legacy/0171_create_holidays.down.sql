-- Drop indexes
DROP INDEX IF EXISTS idx_holidays_semester_id;
DROP INDEX IF EXISTS idx_holidays_start_date;
DROP INDEX IF EXISTS idx_holidays_end_date;
DROP INDEX IF EXISTS idx_holidays_type;

-- Remove foreign key constraint
ALTER TABLE holidays DROP CONSTRAINT IF EXISTS fk_holidays_semester_id;

-- Drop holidays table
DROP TABLE IF EXISTS holidays;
