ALTER TABLE homework_users
    DROP COLUMN IF EXISTS rate;

ALTER TABLE exam_users
    DROP COLUMN IF EXISTS rate;

ALTER TABLE exercise_users
    DROP COLUMN IF EXISTS rate;

ALTER TABLE contest_round_users
    DROP COLUMN IF EXISTS rate;
