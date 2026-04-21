ALTER TABLE cloned_questions
    ADD COLUMN source_questions JSONB NOT NULL DEFAULT '[]';
