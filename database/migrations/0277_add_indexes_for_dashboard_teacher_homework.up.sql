DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_ref_lessons') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_ref_lessons_lesson_course ON homework_ref_lessons (lesson_id, course_id);
        CREATE INDEX IF NOT EXISTS idx_homework_ref_lessons_homework_course ON homework_ref_lessons (homework_id, course_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_users') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_users_homework_user ON homework_users (homework_id, user_id);
        CREATE INDEX IF NOT EXISTS idx_homework_users_user_homework ON homework_users (user_id, homework_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='user_courses') THEN
        CREATE INDEX IF NOT EXISTS idx_user_courses_course_user ON user_courses (course_id, user_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='lesson_schedules') THEN
        CREATE INDEX IF NOT EXISTS idx_lesson_schedules_course_lesson ON lesson_schedules (course_id, lesson_id);
    END IF;
END $$;
