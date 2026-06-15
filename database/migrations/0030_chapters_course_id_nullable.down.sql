-- Cannot restore NOT NULL if nulls exist; best-effort down.
UPDATE chapters SET course_id = 0 WHERE course_id IS NULL;
ALTER TABLE chapters ALTER COLUMN course_id SET NOT NULL;
