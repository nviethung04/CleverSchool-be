-- +migrate Down

ALTER TABLE chapters
DROP COLUMN IF EXISTS "time",
DROP COLUMN IF EXISTS "target";

