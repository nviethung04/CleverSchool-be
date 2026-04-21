CREATE INDEX IF NOT EXISTS idx_homework_ref_lessons_lesson_course
    ON homework_ref_lessons (lesson_id, course_id);

CREATE INDEX IF NOT EXISTS idx_homework_ref_lessons_homework_course
    ON homework_ref_lessons (homework_id, course_id);

CREATE INDEX IF NOT EXISTS idx_homework_users_homework_user
    ON homework_users (homework_id, user_id);

CREATE INDEX IF NOT EXISTS idx_homework_users_user_homework
    ON homework_users (user_id, homework_id);

CREATE INDEX IF NOT EXISTS idx_user_courses_course_user
    ON user_courses (course_id, user_id);

CREATE INDEX IF NOT EXISTS idx_lesson_schedules_course_lesson
    ON lesson_schedules (course_id, lesson_id);

