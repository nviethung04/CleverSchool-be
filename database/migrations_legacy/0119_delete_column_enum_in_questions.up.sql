ALTER TABLE questions ALTER COLUMN "skill" TYPE text;
ALTER TABLE questions ALTER COLUMN "level" TYPE text;
ALTER TABLE questions ALTER COLUMN "topic" TYPE text;

ALTER TABLE questions DROP COLUMN IF EXISTS "skill";
ALTER TABLE questions DROP COLUMN IF EXISTS "level";
ALTER TABLE questions DROP COLUMN IF EXISTS "topic";
