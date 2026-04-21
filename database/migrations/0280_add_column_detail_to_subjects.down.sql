-- +migrate Down

ALTER TABLE subjects
DROP COLUMN IF EXISTS "detail";

