-- homework_users
CREATE INDEX IF NOT EXISTS idx_homework_users_homework_id
    ON homework_users (homework_id);

-- homework_question_users
CREATE INDEX IF NOT EXISTS idx_homework_question_users_homework_id
    ON homework_question_users (homework_id);

-- homework_question_user_positions
CREATE INDEX IF NOT EXISTS idx_homework_question_user_positions_homework_id
    ON homework_question_user_positions (homework_id);

-- homework_question_user_matchings
CREATE INDEX IF NOT EXISTS idx_homework_question_user_matchings_homework_id
    ON homework_question_user_matchings (homework_id);

-- homework_question_user_manual_scoring
CREATE INDEX IF NOT EXISTS idx_homework_question_user_manual_scoring_homework_id
    ON homework_question_user_manual_scoring (homework_id);

-- homework_question_user_labelings
CREATE INDEX IF NOT EXISTS idx_homework_question_user_labelings_homework_id
    ON homework_question_user_labelings (homework_id);

-- homework_question_user_groups
CREATE INDEX IF NOT EXISTS idx_homework_question_user_groups_homework_id
    ON homework_question_user_groups (homework_id);

-- homework_question_user_fill_in_blanks
CREATE INDEX IF NOT EXISTS idx_homework_question_user_fill_in_blanks_homework_id
    ON homework_question_user_fill_in_blanks (homework_id);
