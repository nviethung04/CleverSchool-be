ALTER TABLE courses
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN status TYPE smallint USING CASE
        WHEN status = true THEN 1
        ELSE 0
    END,
    ALTER COLUMN status SET DEFAULT 1;

ALTER TABLE classes
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN status TYPE smallint USING CASE
        WHEN status = true THEN 1
        ELSE 0
    END,
    ALTER COLUMN status SET DEFAULT 1;