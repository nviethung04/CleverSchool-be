ALTER TABLE exam_ref_lessons ADD COLUMN IF NOT EXISTS course_id bigint;

ALTER TABLE homework_ref_lessons ADD COLUMN IF NOT EXISTS course_id bigint;

ALTER TABLE exercise_ref_lessons ADD COLUMN IF NOT EXISTS course_id bigint;

ALTER TABLE lesson_plan_ref_lessons ADD COLUMN IF NOT EXISTS course_id bigint;
