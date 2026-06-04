-- Assessment user summary tables
CREATE TABLE exam_users (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score NUMERIC(5,2),
    ratio NUMERIC(7,2),
    "time" BIGINT,
    has_manual_scoring BOOLEAN DEFAULT FALSE,
    file_infos JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_exam_users_exam_id ON exam_users(exam_id);

CREATE TABLE homework_users (
    id BIGSERIAL PRIMARY KEY,
    homework_id BIGINT NOT NULL REFERENCES homeworks(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_question_id_completed BIGINT DEFAULT 0,
    questions_completed BIGINT DEFAULT 0,
    score NUMERIC(5,2) DEFAULT 0,
    ratio NUMERIC(7,2) DEFAULT 0,
    has_manual_scoring BOOLEAN DEFAULT FALSE,
    status_scoring SMALLINT DEFAULT 0,
    file_infos JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_homework_users_homework_id ON homework_users(homework_id);
CREATE INDEX idx_homework_users_status_scoring ON homework_users(status_scoring);

CREATE TABLE exercise_users (
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score NUMERIC(5,2),
    ratio NUMERIC(7,2),
    "time" BIGINT,
    has_manual_scoring BOOLEAN DEFAULT FALSE,
    file_infos JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_exercise_users_exercise_id ON exercise_users(exercise_id);

-- Macro-like: scoring detail tables per assessment type
-- Exam
CREATE TABLE exam_question_users (id BIGSERIAL PRIMARY KEY, exam_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_id BIGINT, true_answer_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exam_question_user_fill_in_blanks (id BIGSERIAL PRIMARY KEY, exam_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer TEXT, sort_position INT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exam_question_user_positions (id BIGSERIAL PRIMARY KEY, exam_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_group_position BIGINT, sort_position INT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exam_question_user_matchings (id BIGSERIAL PRIMARY KEY, exam_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, first_item_id BIGINT, second_item_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exam_question_user_labelings (id BIGSERIAL PRIMARY KEY, exam_id BIGINT, lesson_id BIGINT, question_id BIGINT, user_id BIGINT, blank_id BIGINT, answer_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exam_question_user_groups (id BIGSERIAL PRIMARY KEY, exam_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_id BIGINT, group_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exam_question_user_manual_scoring (id BIGSERIAL PRIMARY KEY, exam_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer TEXT, file_info JSONB, score NUMERIC(5,2), is_scored BOOLEAN DEFAULT FALSE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, scoring_by BIGINT, scoring_at TIMESTAMP);

-- Homework
CREATE TABLE homework_question_users (id BIGSERIAL PRIMARY KEY, homework_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_id BIGINT, true_answer_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE homework_question_user_fill_in_blanks (id BIGSERIAL PRIMARY KEY, homework_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer TEXT, sort_position INT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE homework_question_user_positions (id BIGSERIAL PRIMARY KEY, homework_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_group_position BIGINT, sort_position INT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE homework_question_user_matchings (id BIGSERIAL PRIMARY KEY, homework_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, first_item_id BIGINT, second_item_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE homework_question_user_labelings (id BIGSERIAL PRIMARY KEY, homework_id BIGINT, lesson_id BIGINT, question_id BIGINT, user_id BIGINT, blank_id BIGINT, answer_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE homework_question_user_groups (id BIGSERIAL PRIMARY KEY, homework_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_id BIGINT, group_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE homework_question_user_manual_scoring (id BIGSERIAL PRIMARY KEY, homework_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer TEXT, file_info JSONB, score NUMERIC(5,2), is_scored BOOLEAN DEFAULT FALSE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, scoring_by BIGINT, scoring_at TIMESTAMP);

-- Exercise
CREATE TABLE exercise_question_users (id BIGSERIAL PRIMARY KEY, exercise_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_id BIGINT, true_answer_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exercise_question_user_fill_in_blanks (id BIGSERIAL PRIMARY KEY, exercise_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer TEXT, sort_position INT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exercise_question_user_positions (id BIGSERIAL PRIMARY KEY, exercise_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_group_position BIGINT, sort_position INT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exercise_question_user_matchings (id BIGSERIAL PRIMARY KEY, exercise_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, first_item_id BIGINT, second_item_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exercise_question_user_labelings (id BIGSERIAL PRIMARY KEY, exercise_id BIGINT, lesson_id BIGINT, question_id BIGINT, user_id BIGINT, blank_id BIGINT, answer_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exercise_question_user_groups (id BIGSERIAL PRIMARY KEY, exercise_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer_id BIGINT, group_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE exercise_question_user_manual_scoring (id BIGSERIAL PRIMARY KEY, exercise_id BIGINT, lesson_id BIGINT, user_id BIGINT, question_id BIGINT, answer TEXT, file_info JSONB, score NUMERIC(5,2), is_scored BOOLEAN DEFAULT FALSE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, scoring_by BIGINT, scoring_at TIMESTAMP);

CREATE TABLE homework_user_skip_questions (
    id BIGSERIAL PRIMARY KEY,
    homework_id BIGINT NOT NULL,
    lesson_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    question_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Level test (optional module, kept for existing code)
CREATE TABLE level_standards (id BIGSERIAL PRIMARY KEY, name BIGINT, create_by BIGINT, create_at TIMESTAMP, update_by BIGINT, update_at TIMESTAMP, deleted_by BIGINT, deleted_at TIMESTAMP);
CREATE TABLE levels (id BIGSERIAL PRIMARY KEY, level_standard_id BIGINT, name BIGINT, min_point NUMERIC(5,2), max_point NUMERIC(5,2), created_by BIGINT, created_at TIMESTAMP, updated_by BIGINT, updated_at TIMESTAMP, deleted_by BIGINT, deleted_at TIMESTAMP);
CREATE TABLE level_tests (id BIGSERIAL PRIMARY KEY, name VARCHAR(50), course_id BIGINT, level_standard_id BIGINT, status SMALLINT, time_limit BIGINT, max_score NUMERIC(5,2), description TEXT, cover_image TEXT, deadline TIMESTAMP, created_at TIMESTAMP, created_by BIGINT, updated_at TIMESTAMP, updated_by BIGINT, deleted_at TIMESTAMP, deleted_by BIGINT);
CREATE TABLE level_test_questions (id BIGSERIAL PRIMARY KEY, level_test_id BIGINT, question_id BIGINT, is_source_question BOOLEAN, score NUMERIC(5,2));
CREATE TABLE level_test_question_users (id BIGSERIAL PRIMARY KEY, level_test_id BIGINT, user_id BIGINT, question_id BIGINT, answer_id BIGINT, true_answer_id BIGINT, is_correct BOOLEAN, score NUMERIC(5,2), created_at TIMESTAMP);
CREATE TABLE level_users (id BIGSERIAL PRIMARY KEY, user_id BIGINT, level_test_id BIGINT, score NUMERIC(5,2), level_id BIGINT, created_at TIMESTAMP);
