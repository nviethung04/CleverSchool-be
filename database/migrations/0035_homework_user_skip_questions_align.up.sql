-- Align homework_user_skip_questions with models/homework_user_skip_question.go
-- (baseline 0009 was missing skip/redo tracking columns).

ALTER TABLE homework_user_skip_questions
    ADD COLUMN IF NOT EXISTS did_it_again BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS did_it_again_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_homework_user_skip_questions_did_it_again
    ON homework_user_skip_questions (did_it_again);
