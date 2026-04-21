-- +migrate Down

DROP INDEX IF EXISTS idx_classes_main_deleted_at;
DROP INDEX IF EXISTS idx_classes_main_school_id;

DROP TABLE IF EXISTS classes_main;

DROP SEQUENCE IF EXISTS classes_main_id_seq;

