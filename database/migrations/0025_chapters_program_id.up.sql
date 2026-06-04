-- chapters.program_id: denormalized field used by exam/homework/dashboard joins (see models/chapter.go).
-- Baseline 0002 only had course_id; backfill from parent course.

ALTER TABLE chapters
    ADD COLUMN IF NOT EXISTS program_id BIGINT REFERENCES programs(id) ON DELETE SET NULL;

UPDATE chapters ch
SET program_id = c.program_id
FROM courses c
WHERE c.id = ch.course_id
  AND (ch.program_id IS NULL OR ch.program_id IS DISTINCT FROM c.program_id);

CREATE INDEX IF NOT EXISTS idx_chapters_program_id ON chapters(program_id);
