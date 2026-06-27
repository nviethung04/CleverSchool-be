-- Align group_answers / answer_groups with Go models (category question type).
-- Greenfield 0006 inverted parent/child; code expects group_answers = categories, answer_groups = items.

-- group_answers: category headers (danh mục)
ALTER TABLE group_answers ADD COLUMN IF NOT EXISTS kind kind_enum;

UPDATE group_answers SET kind = 'text'::kind_enum WHERE kind IS NULL;

ALTER TABLE group_answers ALTER COLUMN kind SET DEFAULT 'text';
ALTER TABLE group_answers ALTER COLUMN kind SET NOT NULL;

ALTER TABLE group_answers DROP CONSTRAINT IF EXISTS group_answers_answer_group_id_fkey;
ALTER TABLE group_answers DROP COLUMN IF EXISTS answer_group_id;
ALTER TABLE group_answers DROP COLUMN IF EXISTS is_correct;

-- answer_groups: items assigned to a category
ALTER TABLE answer_groups ADD COLUMN IF NOT EXISTS group_id BIGINT;
ALTER TABLE answer_groups ADD COLUMN IF NOT EXISTS content TEXT;
ALTER TABLE answer_groups ADD COLUMN IF NOT EXISTS file_info JSONB;
ALTER TABLE answer_groups ADD COLUMN IF NOT EXISTS kind kind_enum;
ALTER TABLE answer_groups ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) DEFAULT 0;

UPDATE answer_groups
SET content = name
WHERE content IS NULL AND name IS NOT NULL;

UPDATE answer_groups SET kind = 'text'::kind_enum WHERE kind IS NULL;

ALTER TABLE answer_groups ALTER COLUMN kind SET DEFAULT 'text';
ALTER TABLE answer_groups ALTER COLUMN kind SET NOT NULL;

ALTER TABLE answer_groups DROP COLUMN IF EXISTS name;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'answer_groups_group_id_fkey'
    ) THEN
        ALTER TABLE answer_groups
            ADD CONSTRAINT answer_groups_group_id_fkey
            FOREIGN KEY (group_id) REFERENCES group_answers(id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_answer_groups_question_id ON answer_groups(question_id);
CREATE INDEX IF NOT EXISTS idx_answer_groups_group_id ON answer_groups(group_id);
