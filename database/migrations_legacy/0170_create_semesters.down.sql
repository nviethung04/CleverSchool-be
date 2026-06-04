-- Drop course_ref_semesters table
DROP TABLE IF EXISTS course_ref_semesters;

-- Drop indexes
DROP INDEX IF EXISTS idx_course_ref_semesters_course_id;
DROP INDEX IF EXISTS idx_course_ref_semesters_semester_id;

-- Drop semesters table
DROP TABLE IF EXISTS semesters;
