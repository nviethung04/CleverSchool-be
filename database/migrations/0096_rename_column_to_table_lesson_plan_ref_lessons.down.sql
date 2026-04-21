ALTER TABLE lesson_plan_ref_lessons
RENAME COLUMN lesson_plan_id TO lesson_plans_id;

ALTER TABLE lesson_plan_ref_lessons
RENAME COLUMN lesson_id TO lessons_id;
