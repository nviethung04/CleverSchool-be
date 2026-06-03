ALTER TABLE IF EXISTS homework_question_user_manual_scoring
ADD COLUMN IF NOT EXISTS scoring_by BIGINT,
ADD COLUMN IF NOT EXISTS scoring_at TIMESTAMP;


