DROP INDEX IF EXISTS idx_question_ref_attributes_parent;

ALTER TABLE question_ref_attributes
    DROP COLUMN IF EXISTS weight,
    DROP COLUMN IF EXISTS parent_attribute_id;
