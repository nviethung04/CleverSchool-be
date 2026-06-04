UPDATE lesson_plans lp
SET lesson_id = lprl.lesson_id
FROM lesson_plan_ref_lessons lprl
WHERE lp.id = lprl.lesson_plan_id;
