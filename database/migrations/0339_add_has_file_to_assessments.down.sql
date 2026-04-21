-- +migrate Down

ALTER TABLE assessments
    DROP COLUMN IF EXISTS has_file;


