ALTER TABLE homework_users
    ADD COLUMN IF NOT EXISTS rate VARCHAR(50) DEFAULT 'not_rated_yet';

ALTER TABLE exam_users
    ADD COLUMN IF NOT EXISTS rate VARCHAR(50) DEFAULT 'not_rated_yet';

ALTER TABLE exercise_users
    ADD COLUMN IF NOT EXISTS rate VARCHAR(50) DEFAULT 'not_rated_yet';

ALTER TABLE contest_round_users
    ADD COLUMN IF NOT EXISTS rate VARCHAR(50) DEFAULT 'not_rated_yet';
