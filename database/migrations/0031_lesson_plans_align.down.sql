DROP INDEX IF EXISTS idx_lesson_plans_sort_position;
DROP INDEX IF EXISTS idx_lesson_plans_program_id;

ALTER TABLE lesson_plans
    DROP COLUMN IF EXISTS views,
    DROP COLUMN IF EXISTS total_time,
    DROP COLUMN IF EXISTS sort_position,
    DROP COLUMN IF EXISTS cover_image_info,
    DROP COLUMN IF EXISTS object_title,
    DROP COLUMN IF EXISTS program_id;
