ALTER TABLE classes
    DROP COLUMN IF EXISTS current_students,
    DROP COLUMN IF EXISTS max_students;
