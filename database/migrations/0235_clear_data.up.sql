UPDATE chapters
SET deleted_at = NOW(),
    deleted_by = 1
WHERE program_id = 0
   OR program_id IS NULL;

UPDATE lessons l
SET deleted_at = NOW(),
    deleted_by = 1
WHERE l.chapter_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM chapters c
      WHERE c.id = l.chapter_id
        AND c.deleted_at IS NULL
  );

UPDATE lesson_plans lp
SET deleted_at = NOW(),
    deleted_by = 1
WHERE NOT EXISTS (
    SELECT 1
    FROM lesson_plan_ref_lessons lprl
    LEFT JOIN lessons l ON lprl.lesson_id = l.id
    WHERE lprl.lesson_plan_id = lp.id
      AND l.deleted_at IS NULL
);

UPDATE homeworks h
SET deleted_at = NOW(),
    deleted_by = 1
WHERE NOT EXISTS (
    SELECT 1
    FROM homework_ref_lessons hrl
    JOIN lessons l ON hrl.lesson_id = l.id
    WHERE hrl.homework_id = h.id
      AND l.deleted_at IS NULL
);

UPDATE cloned_questions cq
SET deleted_at = NOW(),
    deleted_by = 1
WHERE cq.assignment_type = 'homework'
  AND NOT EXISTS (
      SELECT 1
      FROM homeworks h
      WHERE h.id = cq.assignment_id
        AND h.deleted_at IS NULL
);
