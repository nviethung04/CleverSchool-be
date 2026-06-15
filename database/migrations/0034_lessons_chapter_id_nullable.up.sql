-- Lessons can be created in the bank first, then linked to a chapter later.
ALTER TABLE lessons
    ALTER COLUMN chapter_id DROP NOT NULL;
