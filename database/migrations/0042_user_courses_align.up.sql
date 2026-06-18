-- Align user_courses with models/user_course.go (code expects main_teacher, not is_main_teacher).

ALTER TABLE user_courses
    ADD COLUMN IF NOT EXISTS main_teacher BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_current BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS start_time TIMESTAMP,
    ADD COLUMN IF NOT EXISTS end_time TIMESTAMP;

UPDATE user_courses
SET main_teacher = is_main_teacher
WHERE is_main_teacher = TRUE;
