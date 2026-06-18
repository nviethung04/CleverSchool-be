ALTER TABLE user_courses
    DROP COLUMN IF EXISTS main_teacher,
    DROP COLUMN IF EXISTS is_current,
    DROP COLUMN IF EXISTS start_time,
    DROP COLUMN IF EXISTS end_time;
