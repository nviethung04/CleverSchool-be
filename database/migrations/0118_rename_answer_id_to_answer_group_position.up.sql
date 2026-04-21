-- Rename column in exam_question_user_positions
ALTER TABLE exam_question_user_positions
    RENAME COLUMN answer_id TO answer_group_position;

-- Rename column in homework_question_user_positions
ALTER TABLE homework_question_user_positions
    RENAME COLUMN answer_id TO answer_group_position;
