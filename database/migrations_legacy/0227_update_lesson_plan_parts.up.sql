ALTER TABLE lesson_plan_parts ADD COLUMN course_id BIGINT;
CREATE INDEX idx_lesson_plan_parts_course_id ON lesson_plan_parts(course_id);
