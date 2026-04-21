CREATE INDEX IF NOT EXISTS idx_user_classes_class_id
    ON user_classes (class_id);

CREATE INDEX IF NOT EXISTS idx_user_classes_user_id
    ON user_classes (user_id);

CREATE INDEX IF NOT EXISTS idx_user_courses_class_id
    ON user_courses (course_id);

CREATE INDEX IF NOT EXISTS idx_user_courses_user_id
    ON user_courses (user_id);
