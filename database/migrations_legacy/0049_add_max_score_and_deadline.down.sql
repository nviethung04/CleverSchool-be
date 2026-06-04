-- Xoá cột 'max_score' khỏi bảng 'lesson_plan_parts' nếu tồn tại
ALTER TABLE public.lesson_plan_parts
DROP COLUMN IF EXISTS max_score;

-- Xoá cột 'deadline' khỏi bảng 'exams' nếu tồn tại
ALTER TABLE public.exams
DROP COLUMN IF EXISTS deadline;
