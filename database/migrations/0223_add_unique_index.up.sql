ALTER TABLE homework_ref_lessons
ADD CONSTRAINT homework_ref_lessons_unique
UNIQUE (homework_id, lesson_id, course_id);

ALTER TABLE exercise_ref_lessons
ADD CONSTRAINT exercise_ref_lessons_unique
UNIQUE (exercise_id, lesson_id, course_id);

ALTER TABLE exam_ref_lessons
ADD CONSTRAINT exam_ref_lessons_unique
UNIQUE (exam_id, lesson_id, course_id);
