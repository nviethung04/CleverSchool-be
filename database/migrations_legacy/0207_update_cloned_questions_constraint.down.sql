-- Revert cloned_questions constraint to exclude 'exercise' assignment type
ALTER TABLE cloned_questions 
DROP CONSTRAINT IF EXISTS cloned_questions_assignment_type_check;

ALTER TABLE cloned_questions 
ADD CONSTRAINT cloned_questions_assignment_type_check 
CHECK (assignment_type IN ('homework', 'exam', 'lesson_plan_part', 'level_test'));
