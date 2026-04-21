ALTER TABLE user_courses
RENAME COLUMN user_id TO users_id;

ALTER TABLE user_courses
RENAME COLUMN course_id TO courses_id;

ALTER TABLE user_courses
RENAME TO users_courses;
