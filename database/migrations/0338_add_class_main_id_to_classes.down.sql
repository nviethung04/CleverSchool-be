-- +migrate Down

DROP INDEX IF EXISTS idx_classes_class_main_id;

ALTER TABLE classes DROP COLUMN IF EXISTS class_main_id;

