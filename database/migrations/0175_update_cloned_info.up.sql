-- Course
UPDATE courses c
SET clone_info = c.clone_info || jsonb_build_object('program_id', p.id)
FROM programs p
WHERE p.name = c.name
  AND c.clone_info IS NOT NULL;

UPDATE courses c
SET clone_info = jsonb_build_object(
        'clone_id', 0,
        'cloned_at', '',
        'course_id', 0,
        'program_id', p.id
    )
FROM programs p
WHERE p.name = c.name
  AND c.clone_info IS NULL;

-- Chapter
UPDATE chapters ch
SET clone_info = ch.clone_info || jsonb_build_object(
        'program_id', COALESCE((c.clone_info->>'program_id')::INT, 0)
    )
FROM courses c
WHERE c.id = ch.course_id
  AND ch.clone_info IS NOT NULL;

UPDATE chapters ch
SET clone_info = jsonb_build_object(
        'clone_id', 0,
        'cloned_at', '',
        'course_id', ch.course_id,
        'program_id', COALESCE((c.clone_info->>'program_id')::INT, 0)
    )
FROM courses c
WHERE c.id = ch.course_id
  AND ch.clone_info IS NULL;

-- Lesson
UPDATE lessons l
SET clone_info = l.clone_info || jsonb_build_object(
        'program_id', COALESCE((ch.clone_info->>'program_id')::INT, 0)
    )
FROM chapters ch
WHERE ch.id = l.chapter_id
  AND l.clone_info IS NOT NULL;

UPDATE lessons l
SET clone_info = jsonb_build_object(
        'clone_id', 0,
        'cloned_at', '',
        'lesson_id', l.id,
        'program_id', COALESCE((ch.clone_info->>'program_id')::INT, 0)
    )
FROM chapters ch
WHERE ch.id = l.chapter_id
  AND l.clone_info IS NULL;


-- LessonPlan
UPDATE lesson_plans lp
SET clone_info = lp.clone_info || jsonb_build_object(
        'program_id', COALESCE((l.clone_info->>'program_id')::INT, 0)
    )
FROM lesson_plan_ref_lessons lprl
JOIN lessons l ON l.id = lprl.lesson_id
WHERE lprl.lesson_plan_id = lp.id
  AND lp.clone_info IS NOT NULL;

UPDATE lesson_plans lp
SET clone_info = jsonb_build_object(
        'clone_id', 0,
        'cloned_at', '',
        'lesson_plan_id', lp.id,
        'program_id', COALESCE((l.clone_info->>'program_id')::INT, 0)
    )
FROM lesson_plan_ref_lessons lprl
JOIN lessons l ON l.id = lprl.lesson_id
WHERE lprl.lesson_plan_id = lp.id
  AND lp.clone_info IS NULL;

-- LessonPlanPart
UPDATE lesson_plan_parts lpp
SET clone_info = lpp.clone_info || jsonb_build_object(
        'program_id', COALESCE((lp.clone_info->>'program_id')::INT, 0)
    )
FROM lesson_plans lp
WHERE lp.id = lpp.lesson_plan_id
  AND lpp.clone_info IS NOT NULL;

UPDATE lesson_plan_parts lpp
SET clone_info = jsonb_build_object(
        'clone_id', 0,
        'cloned_at', '',
        'lesson_plan_part_id', lpp.id,
        'program_id', COALESCE((lp.clone_info->>'program_id')::INT, 0)
    )
FROM lesson_plans lp
WHERE lp.id = lpp.lesson_plan_id
  AND lpp.clone_info IS NULL;

-- Exam
UPDATE exams e
SET clone_info = e.clone_info || jsonb_build_object(
        'program_id', COALESCE((l.clone_info->>'program_id')::INT, 0)
    )
FROM lessons l
WHERE l.id = e.lesson_id
  AND e.clone_info IS NOT NULL;

UPDATE exams e
SET clone_info = jsonb_build_object(
        'clone_id', 0,
        'cloned_at', '',
        'exam_id', e.id,
        'program_id', COALESCE((l.clone_info->>'program_id')::INT, 0)
    )
FROM lessons l
WHERE l.id = e.lesson_id
  AND e.clone_info IS NULL;

-- Homework
UPDATE homeworks h
SET clone_info = h.clone_info || jsonb_build_object(
        'program_id', COALESCE((l.clone_info->>'program_id')::INT, 0)
    )
FROM lessons l
WHERE l.id = h.lesson_id
  AND h.clone_info IS NOT NULL;

UPDATE homeworks h
SET clone_info = jsonb_build_object(
        'clone_id', 0,
        'cloned_at', '',
        'homework_id', h.id,
        'program_id', COALESCE((l.clone_info->>'program_id')::INT, 0)
    )
FROM lessons l
WHERE l.id = h.lesson_id
  AND h.clone_info IS NULL;

-- ClonedQuestion
-- Cập nhật clone_info cho homework
UPDATE cloned_questions cq
SET clone_info = COALESCE(cq.clone_info, '{}'::jsonb) ||
    jsonb_build_object(
        'clone_id', COALESCE((cq.clone_info->>'clone_id')::INT, 0),
        'cloned_at', COALESCE(cq.clone_info->>'cloned_at', ''),
        'homework_id', assignment_id,
        'program_id', COALESCE((h.clone_info->>'program_id')::INT, 0)
    )
FROM homeworks h
WHERE cq.assignment_type = 'homework'
  AND cq.assignment_id = h.id;

-- Cập nhật clone_info cho exam
UPDATE cloned_questions cq
SET clone_info = COALESCE(cq.clone_info, '{}'::jsonb) ||
    jsonb_build_object(
        'clone_id', COALESCE((cq.clone_info->>'clone_id')::INT, 0),
        'cloned_at', COALESCE(cq.clone_info->>'cloned_at', ''),
        'exam_id', assignment_id,
        'program_id', COALESCE((e.clone_info->>'program_id')::INT, 0)
    )
FROM exams e
WHERE cq.assignment_type = 'exam'
  AND cq.assignment_id = e.id;

