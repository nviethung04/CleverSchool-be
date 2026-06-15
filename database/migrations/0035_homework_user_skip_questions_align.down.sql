DROP INDEX IF EXISTS idx_homework_user_skip_questions_did_it_again;

ALTER TABLE homework_user_skip_questions
    DROP COLUMN IF EXISTS did_it_again,
    DROP COLUMN IF EXISTS did_it_again_at,
    DROP COLUMN IF EXISTS updated_at;
