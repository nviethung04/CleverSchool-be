ALTER TABLE question_ref_attributes
    ADD COLUMN IF NOT EXISTS parent_attribute_id BIGINT REFERENCES question_attributes(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS weight NUMERIC(10, 2) DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_question_ref_attributes_parent
    ON question_ref_attributes(parent_attribute_id);
