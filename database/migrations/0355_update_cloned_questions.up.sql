ALTER TABLE cloned_questions
    ADD COLUMN sort_question_ids JSONB NOT NULL DEFAULT '[]';
