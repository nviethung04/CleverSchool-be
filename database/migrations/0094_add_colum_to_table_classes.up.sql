ALTER TABLE classes
    ADD COLUMN IF NOT EXISTS current_students INT,
    ADD COLUMN IF NOT EXISTS max_students INT;
