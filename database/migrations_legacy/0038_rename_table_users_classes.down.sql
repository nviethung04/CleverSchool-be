ALTER TABLE user_classes
RENAME TO users_classes;

ALTER TABLE users_classes
RENAME COLUMN user_id TO users_id;

ALTER TABLE users_classes
RENAME COLUMN class_id TO classes_id;
