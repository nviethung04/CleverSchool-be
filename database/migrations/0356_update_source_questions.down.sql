ALTER TABLE source_questions
    DROP COLUMN IF EXISTS file_infos;

-- Note: PostgreSQL does not support removing enum values directly.
-- The 'default' values added to skill_enum and level_enum will remain,
-- but this should not cause issues as they are valid enum values.
