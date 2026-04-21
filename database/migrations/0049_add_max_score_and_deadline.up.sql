-- Thêm cột 'max_score' vào bảng 'lesson_plan_parts'
ALTER TABLE public.lesson_plan_parts
    ADD COLUMN IF NOT EXISTS max_score numeric(5,2);

-- Thêm cột 'deadline' vào bảng 'exams'
ALTER TABLE public.exams
    ADD COLUMN IF NOT EXISTS deadline timestamp without time zone;
