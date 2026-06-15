DROP INDEX IF EXISTS idx_lesson_plan_parts_course_id;
DROP INDEX IF EXISTS idx_lesson_plan_parts_program_id;

ALTER TABLE lesson_plan_parts
    DROP COLUMN IF EXISTS max_score,
    DROP COLUMN IF EXISTS file,
    DROP COLUMN IF EXISTS guide_student,
    DROP COLUMN IF EXISTS guide_teacher,
    DROP COLUMN IF EXISTS link_type,
    DROP COLUMN IF EXISTS link_info,
    DROP COLUMN IF EXISTS file_type,
    DROP COLUMN IF EXISTS is_classwork,
    DROP COLUMN IF EXISTS time,
    DROP COLUMN IF EXISTS tag,
    DROP COLUMN IF EXISTS course_id,
    DROP COLUMN IF EXISTS program_id;
