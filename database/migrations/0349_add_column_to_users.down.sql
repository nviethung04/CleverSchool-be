ALTER TABLE exams
    DROP COLUMN IF EXISTS is_independent_student,
    DROP COLUMN IF EXISTS is_failed_subject;
