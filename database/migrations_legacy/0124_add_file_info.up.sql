CREATE TABLE IF NOT EXISTS exam_question_user_manual_scoring (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT,
    lesson_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer TEXT,
    file_info JSONB,
    score NUMERIC(5, 2),
    is_scored BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    scoring_by BIGINT,
    scoring_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS homework_question_user_manual_scoring (
    id BIGSERIAL PRIMARY KEY,
    homework_id BIGINT,
    lesson_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer TEXT,
    file_info JSONB,
    score NUMERIC(5, 2),
    is_scored BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    scoring_by BIGINT,
    scoring_at TIMESTAMP
);

ALTER TABLE exam_question_user_manual_scoring
ADD COLUMN IF NOT EXISTS exam_id BIGINT,
ADD COLUMN IF NOT EXISTS lesson_id BIGINT,
ADD COLUMN IF NOT EXISTS user_id BIGINT,
ADD COLUMN IF NOT EXISTS question_id BIGINT,
ADD COLUMN IF NOT EXISTS answer TEXT,
ADD COLUMN IF NOT EXISTS score NUMERIC(5, 2),
ADD COLUMN IF NOT EXISTS is_scored BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS scoring_by BIGINT,
ADD COLUMN IF NOT EXISTS scoring_at TIMESTAMP;

ALTER TABLE homework_question_user_manual_scoring
ADD COLUMN IF NOT EXISTS homework_id BIGINT,
ADD COLUMN IF NOT EXISTS lesson_id BIGINT,
ADD COLUMN IF NOT EXISTS user_id BIGINT,
ADD COLUMN IF NOT EXISTS question_id BIGINT,
ADD COLUMN IF NOT EXISTS answer TEXT,
ADD COLUMN IF NOT EXISTS score NUMERIC(5, 2),
ADD COLUMN IF NOT EXISTS is_scored BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS scoring_by BIGINT,
ADD COLUMN IF NOT EXISTS scoring_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_exam_manual_scoring_user_question
    ON exam_question_user_manual_scoring (exam_id, user_id, question_id);

CREATE INDEX IF NOT EXISTS idx_exam_manual_scoring_unscored
    ON exam_question_user_manual_scoring (exam_id, user_id, is_scored);

CREATE INDEX IF NOT EXISTS idx_exam_manual_scoring_created_at
    ON exam_question_user_manual_scoring (created_at);

CREATE INDEX IF NOT EXISTS idx_homework_manual_scoring_user_question
    ON homework_question_user_manual_scoring (homework_id, user_id, question_id);

CREATE INDEX IF NOT EXISTS idx_homework_manual_scoring_unscored
    ON homework_question_user_manual_scoring (homework_id, user_id, is_scored);

CREATE INDEX IF NOT EXISTS idx_homework_manual_scoring_created_at
    ON homework_question_user_manual_scoring (created_at);

ALTER TABLE users
ADD COLUMN IF NOT EXISTS avatar_info JSONB;

ALTER TABLE certificates
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE courses
ADD COLUMN IF NOT EXISTS image_info JSONB;

ALTER TABLE degrees
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE exam_question_user_manual_scoring
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE exams
ADD COLUMN IF NOT EXISTS cover_image_info JSONB;

ALTER TABLE homework_question_user_manual_scoring
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE homeworks
ADD COLUMN IF NOT EXISTS cover_image_info JSONB;

ALTER TABLE lesson_plan_parts
ADD COLUMN IF NOT EXISTS link_info JSONB;

ALTER TABLE lesson_plans
ADD COLUMN IF NOT EXISTS cover_image_info JSONB;

ALTER TABLE level_tests
ADD COLUMN IF NOT EXISTS cover_image_info JSONB;

ALTER TABLE schools
ADD COLUMN IF NOT EXISTS logo_info JSONB;

ALTER TABLE skills
ADD COLUMN IF NOT EXISTS image_info JSONB;

ALTER TABLE tags
ADD COLUMN IF NOT EXISTS image_info JSONB;

ALTER TABLE topics
ADD COLUMN IF NOT EXISTS image_info JSONB;

ALTER TABLE questions
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE answer_groups
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE answer_coordinates
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE answer_matchings
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE answer_positions
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE answers
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE group_answers
ADD COLUMN IF NOT EXISTS file_info JSONB;

ALTER TABLE answer_matchings
ADD COLUMN IF NOT EXISTS matching_file_info JSONB;

ALTER TABLE questions
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_groups
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_coordinates
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_matchings
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_positions
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answers
DROP COLUMN IF EXISTS media_info;

ALTER TABLE group_answers
DROP COLUMN IF EXISTS media_info;

ALTER TABLE answer_matchings
DROP COLUMN IF EXISTS matching_media_info;
