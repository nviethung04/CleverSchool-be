-- models.LessonPlan.status is int; baseline 0007 used BOOLEAN.

ALTER TABLE lesson_plans
    ALTER COLUMN status DROP DEFAULT;

ALTER TABLE lesson_plans
    ALTER COLUMN status TYPE SMALLINT
    USING CASE
        WHEN status IS TRUE THEN 1
        WHEN status IS FALSE THEN 0
        ELSE 0
    END;

ALTER TABLE lesson_plans
    ALTER COLUMN status SET DEFAULT 1;
