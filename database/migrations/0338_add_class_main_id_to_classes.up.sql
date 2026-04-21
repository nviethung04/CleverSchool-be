-- +migrate Up

ALTER TABLE classes ADD COLUMN IF NOT EXISTS class_main_id bigint;

CREATE INDEX IF NOT EXISTS idx_classes_class_main_id ON classes(class_main_id);

