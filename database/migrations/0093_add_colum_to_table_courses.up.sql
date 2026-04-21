ALTER TABLE courses
    ADD COLUMN IF NOT EXISTS current_students INT;
