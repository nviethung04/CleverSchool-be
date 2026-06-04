ALTER TABLE homework_ref_lessons
DROP CONSTRAINT IF EXISTS homework_ref_lessons_unique;

ALTER TABLE exercise_ref_lessons
DROP CONSTRAINT IF EXISTS exercise_ref_lessons_unique;

ALTER TABLE exam_ref_lessons
DROP CONSTRAINT IF EXISTS exam_ref_lessons_unique;
