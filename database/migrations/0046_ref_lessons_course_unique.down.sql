DROP INDEX IF EXISTS exam_ref_lessons_course_uq;
DROP INDEX IF EXISTS homework_ref_lessons_course_uq;
DROP INDEX IF EXISTS exercise_ref_lessons_course_uq;
DROP INDEX IF EXISTS exam_ref_lessons_program_uq;
DROP INDEX IF EXISTS homework_ref_lessons_program_uq;
DROP INDEX IF EXISTS exercise_ref_lessons_program_uq;

ALTER TABLE exam_ref_lessons
    ADD CONSTRAINT exam_ref_lessons_lesson_id_exam_id_key UNIQUE (lesson_id, exam_id);

ALTER TABLE homework_ref_lessons
    ADD CONSTRAINT homework_ref_lessons_lesson_id_homework_id_key UNIQUE (lesson_id, homework_id);

ALTER TABLE exercise_ref_lessons
    ADD CONSTRAINT exercise_ref_lessons_lesson_id_exercise_id_key UNIQUE (lesson_id, exercise_id);
