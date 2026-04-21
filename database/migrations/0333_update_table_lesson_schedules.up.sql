ALTER TABLE lesson_schedules
  ADD COLUMN IF NOT EXISTS lesson_plan_id BIGINT;
