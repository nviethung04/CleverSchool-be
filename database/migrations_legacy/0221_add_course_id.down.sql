ALTER TABLE exam_ref_lessons DROP COLUMN IF EXISTS course_id;

ALTER TABLE homework_ref_lessons DROP COLUMN IF EXISTS course_id;

ALTER TABLE exercise_ref_lessons DROP COLUMN IF EXISTS course_id;

ALTER TABLE lesson_plan_ref_lessons DROP COLUMN IF EXISTS course_id;
