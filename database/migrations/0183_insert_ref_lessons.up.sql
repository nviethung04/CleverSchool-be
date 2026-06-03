INSERT INTO homework_ref_lessons (homework_id, lesson_id)
SELECT h.id, h.lesson_id
FROM homeworks h
JOIN lessons l ON h.lesson_id = l.id
WHERE h.lesson_id IS NOT NULL
  AND h.lesson_id != 0;

INSERT INTO exam_ref_lessons (exam_id, lesson_id)
SELECT e.id, e.lesson_id
FROM exams e
JOIN lessons l ON e.lesson_id = l.id
WHERE e.lesson_id IS NOT NULL
  AND e.lesson_id != 0;
