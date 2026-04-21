ALTER TABLE users_courses
RENAME COLUMN users_id TO user_id;

ALTER TABLE users_courses
RENAME COLUMN courses_id TO course_id;

ALTER TABLE users_courses
RENAME TO user_courses;
