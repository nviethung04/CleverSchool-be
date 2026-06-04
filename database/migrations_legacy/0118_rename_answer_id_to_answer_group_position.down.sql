-- Rename column back in exam_question_user_positions
ALTER TABLE exam_question_user_positions
    RENAME COLUMN answer_group_position TO answer_id;

-- Rename column back in homework_question_user_positions
ALTER TABLE homework_question_user_positions
    RENAME COLUMN answer_group_position TO answer_id;
