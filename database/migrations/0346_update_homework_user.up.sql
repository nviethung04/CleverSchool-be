ALTER TABLE IF EXISTS homework_users
    ADD COLUMN IF NOT EXISTS rate VARCHAR(50) DEFAULT 'not_rated_yet';

ALTER TABLE IF EXISTS exam_users
    ADD COLUMN IF NOT EXISTS rate VARCHAR(50) DEFAULT 'not_rated_yet';

ALTER TABLE IF EXISTS exercise_users
    ADD COLUMN IF NOT EXISTS rate VARCHAR(50) DEFAULT 'not_rated_yet';

ALTER TABLE IF EXISTS contest_round_users
    ADD COLUMN IF NOT EXISTS rate VARCHAR(50) DEFAULT 'not_rated_yet';

