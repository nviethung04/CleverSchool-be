-- +migrate Down

ALTER TABLE programs
DROP COLUMN IF EXISTS "detail";

