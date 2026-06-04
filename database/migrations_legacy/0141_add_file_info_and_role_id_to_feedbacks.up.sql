ALTER TABLE feedbacks ADD COLUMN IF NOT EXISTS file_info json;
ALTER TABLE feedbacks ADD COLUMN IF NOT EXISTS role_id bigint;
