ALTER TABLE lesson_plans ADD COLUMN lesson_id BIGINT;
CREATE INDEX idx_lesson_plans_lesson_id ON lesson_plans(lesson_id);

DELETE FROM lesson_plan_ref_lessons lprl
WHERE NOT EXISTS (
    SELECT 1
    FROM lessons l
    WHERE l.id = lprl.lesson_id
);
