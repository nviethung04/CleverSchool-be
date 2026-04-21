-- Drop indexes
DROP INDEX IF EXISTS idx_contest_round_joiner_classes_class_id;
DROP INDEX IF EXISTS idx_contest_round_joiner_classes_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_joiner_persons_user_id;
DROP INDEX IF EXISTS idx_contest_round_joiner_persons_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_joiner_provinces_province_id;
DROP INDEX IF EXISTS idx_contest_round_joiner_provinces_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_joiner_schools_school_id;
DROP INDEX IF EXISTS idx_contest_round_joiner_schools_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_fill_in_blanks_question_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_fill_in_blanks_user_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_fill_in_blanks_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_groups_question_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_groups_user_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_groups_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_labelings_question_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_labelings_user_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_labelings_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_manual_scoring_question_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_manual_scoring_user_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_manual_scoring_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_matchings_question_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_matchings_user_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_matchings_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_multiple_choices_question_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_multiple_choices_user_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_multiple_choices_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_positions_question_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_positions_user_id;
DROP INDEX IF EXISTS idx_contest_round_question_user_positions_contest_round_id;
DROP INDEX IF EXISTS idx_contest_round_users_user_id;
DROP INDEX IF EXISTS idx_contest_round_users_contest_round_id;
DROP INDEX IF EXISTS idx_contest_rounds_contest_id;

-- Revert cloned_questions constraint
ALTER TABLE cloned_questions DROP CONSTRAINT IF EXISTS cloned_questions_assignment_type_check;
ALTER TABLE cloned_questions ADD CONSTRAINT cloned_questions_assignment_type_check 
    CHECK (assignment_type = ANY (ARRAY['homework'::text, 'exam'::text, 'lesson_plan_part'::text, 'level_test'::text, 'exercise'::text]));

-- Drop tables in reverse order
DROP TABLE IF EXISTS contest_round_joiner_classes;
DROP TABLE IF EXISTS contest_round_joiner_persons;
DROP TABLE IF EXISTS contest_round_joiner_provinces;
DROP TABLE IF EXISTS contest_round_joiner_schools;
DROP TABLE IF EXISTS contest_round_question_user_fill_in_blanks;
DROP TABLE IF EXISTS contest_round_question_user_groups;
DROP TABLE IF EXISTS contest_round_question_user_labelings;
DROP TABLE IF EXISTS contest_round_question_user_manual_scoring;
DROP TABLE IF EXISTS contest_round_question_user_matchings;
DROP TABLE IF EXISTS contest_round_question_user_multiple_choices;
DROP TABLE IF EXISTS contest_round_question_user_positions;
DROP TABLE IF EXISTS contest_round_users;
DROP TABLE IF EXISTS contest_rounds;
DROP TABLE IF EXISTS contests;

-- Drop enum type
DROP TYPE IF EXISTS join_level_enum;
