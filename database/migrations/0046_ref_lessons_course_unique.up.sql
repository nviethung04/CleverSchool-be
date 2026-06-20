-- Allow program-level (course_id IS NULL) and per-course rows for the same lesson+assignment.
-- Required for POST /homeworks|exams|exercises/:id/assigned (ON CONFLICT lesson_id, *_id, course_id).

ALTER TABLE exam_ref_lessons
    DROP CONSTRAINT IF EXISTS exam_ref_lessons_lesson_id_exam_id_key;

ALTER TABLE homework_ref_lessons
    DROP CONSTRAINT IF EXISTS homework_ref_lessons_lesson_id_homework_id_key;

ALTER TABLE exercise_ref_lessons
    DROP CONSTRAINT IF EXISTS exercise_ref_lessons_lesson_id_exercise_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS exam_ref_lessons_program_uq
    ON exam_ref_lessons (lesson_id, exam_id) WHERE course_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS homework_ref_lessons_program_uq
    ON homework_ref_lessons (lesson_id, homework_id) WHERE course_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS exercise_ref_lessons_program_uq
    ON exercise_ref_lessons (lesson_id, exercise_id) WHERE course_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS exam_ref_lessons_course_uq
    ON exam_ref_lessons (lesson_id, exam_id, course_id) WHERE course_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS homework_ref_lessons_course_uq
    ON homework_ref_lessons (lesson_id, homework_id, course_id) WHERE course_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS exercise_ref_lessons_course_uq
    ON exercise_ref_lessons (lesson_id, exercise_id, course_id) WHERE course_id IS NOT NULL;
