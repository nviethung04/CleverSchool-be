ALTER TABLE courses
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN status TYPE boolean USING CASE
        WHEN status = 1 THEN true
        ELSE false
    END,
    ALTER COLUMN status SET DEFAULT true;

ALTER TABLE classes
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN status TYPE boolean USING CASE
        WHEN status = 1 THEN true
        ELSE false
    END,
    ALTER COLUMN status SET DEFAULT true;
