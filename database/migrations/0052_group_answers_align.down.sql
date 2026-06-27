-- Best-effort rollback (category data in new columns may be lost).

ALTER TABLE answer_groups DROP CONSTRAINT IF EXISTS answer_groups_group_id_fkey;
DROP INDEX IF EXISTS idx_answer_groups_group_id;
DROP INDEX IF EXISTS idx_answer_groups_question_id;

ALTER TABLE answer_groups DROP COLUMN IF EXISTS point;
ALTER TABLE answer_groups DROP COLUMN IF EXISTS kind;
ALTER TABLE answer_groups DROP COLUMN IF EXISTS file_info;
ALTER TABLE answer_groups DROP COLUMN IF EXISTS content;
ALTER TABLE answer_groups DROP COLUMN IF EXISTS group_id;

ALTER TABLE answer_groups ADD COLUMN IF NOT EXISTS name VARCHAR(255);

ALTER TABLE group_answers DROP COLUMN IF EXISTS kind;
ALTER TABLE group_answers ADD COLUMN IF NOT EXISTS answer_group_id BIGINT;
ALTER TABLE group_answers ADD COLUMN IF NOT EXISTS is_correct BOOLEAN DEFAULT FALSE;
