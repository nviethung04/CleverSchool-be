UPDATE lessons SET chapter_id = 0 WHERE chapter_id IS NULL;
ALTER TABLE lessons ALTER COLUMN chapter_id SET NOT NULL;
