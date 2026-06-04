DROP INDEX IF EXISTS idx_chapters_program_id;
ALTER TABLE chapters DROP COLUMN IF EXISTS program_id;
