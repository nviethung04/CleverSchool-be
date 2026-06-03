DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_users') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_users_homework_id ON homework_users (homework_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_question_users') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_question_users_homework_id ON homework_question_users (homework_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_question_user_positions') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_question_user_positions_homework_id ON homework_question_user_positions (homework_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_question_user_matchings') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_question_user_matchings_homework_id ON homework_question_user_matchings (homework_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_question_user_manual_scoring') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_question_user_manual_scoring_homework_id ON homework_question_user_manual_scoring (homework_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_question_user_labelings') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_question_user_labelings_homework_id ON homework_question_user_labelings (homework_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_question_user_groups') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_question_user_groups_homework_id ON homework_question_user_groups (homework_id);
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='homework_question_user_fill_in_blanks') THEN
        CREATE INDEX IF NOT EXISTS idx_homework_question_user_fill_in_blanks_homework_id ON homework_question_user_fill_in_blanks (homework_id);
    END IF;
END $$;
