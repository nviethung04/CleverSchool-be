-- Chapters are program-scoped (see 0025); course_id is legacy from baseline 0002.
ALTER TABLE chapters
    ALTER COLUMN course_id DROP NOT NULL;

UPDATE chapters ch
SET course_id = sub.id
FROM (
    SELECT DISTINCT ON (c.program_id) c.program_id, c.id
    FROM courses c
    WHERE c.deleted_at IS NULL
      AND c.program_id IS NOT NULL
    ORDER BY c.program_id, c.id
) sub
WHERE ch.program_id = sub.program_id
  AND ch.course_id IS NULL;
