ALTER TABLE users_classes
RENAME COLUMN users_id TO user_id;

ALTER TABLE users_classes
RENAME COLUMN classes_id TO class_id;

ALTER TABLE users_classes
RENAME TO user_classes;
