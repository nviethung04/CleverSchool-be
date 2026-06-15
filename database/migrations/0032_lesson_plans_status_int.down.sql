ALTER TABLE lesson_plans
    ALTER COLUMN status DROP DEFAULT;

ALTER TABLE lesson_plans
    ALTER COLUMN status TYPE BOOLEAN
    USING (status <> 0);

ALTER TABLE lesson_plans
    ALTER COLUMN status SET DEFAULT TRUE;
