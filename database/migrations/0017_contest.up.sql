-- join_level_enum đã tạo ở migration 0001

CREATE TABLE IF NOT EXISTS contests (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    status SMALLINT,
    start_time TIMESTAMP WITHOUT TIME ZONE,
    end_time TIMESTAMP WITHOUT TIME ZONE,
    image_info JSONB,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT
);

-- Create contest_rounds table
CREATE TABLE IF NOT EXISTS contest_rounds (
    id BIGSERIAL PRIMARY KEY,
    contest_id BIGINT NOT NULL,
    name TEXT,
    description TEXT,
    sort_position BIGINT,
    join_level join_level_enum,
    start_time TIMESTAMP WITHOUT TIME ZONE,
    end_time TIMESTAMP WITHOUT TIME ZONE,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    updated_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT
);

-- Create contest_round_users table
CREATE TABLE IF NOT EXISTS contest_round_users (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    user_id BIGINT,
    score NUMERIC(5,2),
    ratio NUMERIC(5,2),
    "time" BIGINT,
    has_manual_scoring BOOLEAN,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    updated_at TIMESTAMP WITHOUT TIME ZONE
);

-- Create contest_round_question_user_positions table
CREATE TABLE IF NOT EXISTS contest_round_question_user_positions (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer_group_position BIGINT,
    sort_position INTEGER,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- Create contest_round_question_user_multiple_choices table
CREATE TABLE IF NOT EXISTS contest_round_question_user_multiple_choices (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer_id BIGINT,
    true_answer_id BIGINT,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- Create contest_round_question_user_matchings table
CREATE TABLE IF NOT EXISTS contest_round_question_user_matchings (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    first_item_id BIGINT,
    second_item_id BIGINT,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- Create contest_round_question_user_manual_scoring table
CREATE TABLE IF NOT EXISTS contest_round_question_user_manual_scoring (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer TEXT,
    score NUMERIC(5,2),
    is_scored BOOLEAN,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    scoring_by BIGINT,
    scoring_at TIMESTAMP WITHOUT TIME ZONE,
    file_info JSONB
);

-- Create contest_round_question_user_labelings table
CREATE TABLE IF NOT EXISTS contest_round_question_user_labelings (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    question_id BIGINT,
    user_id BIGINT,
    blank_id BIGINT,
    answer_id BIGINT,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- Create contest_round_question_user_groups table
CREATE TABLE IF NOT EXISTS contest_round_question_user_groups (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer_id BIGINT,
    group_id BIGINT,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- Create contest_round_question_user_fill_in_blanks table
CREATE TABLE IF NOT EXISTS contest_round_question_user_fill_in_blanks (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer TEXT,
    sort_position INTEGER,
    is_correct BOOLEAN,
    score NUMERIC(5,2),
    created_at TIMESTAMP WITHOUT TIME ZONE
);

-- Create contest_round_joiner_schools table
CREATE TABLE IF NOT EXISTS contest_round_joiner_schools (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    school_id BIGINT,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT
);

-- Create contest_round_joiner_provinces table
CREATE TABLE IF NOT EXISTS contest_round_joiner_provinces (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    province_id BIGINT,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT
);

-- Create contest_round_joiner_persons table
CREATE TABLE IF NOT EXISTS contest_round_joiner_persons (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    user_id BIGINT,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT
);

-- Create contest_round_joiner_classes table
CREATE TABLE IF NOT EXISTS contest_round_joiner_classes (
    id BIGSERIAL PRIMARY KEY,
    contest_round_id BIGINT,
    class_id BIGINT,
    created_at TIMESTAMP WITHOUT TIME ZONE,
    created_by BIGINT,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by BIGINT
);

-- Update cloned_questions table to support contest assignment type
ALTER TABLE cloned_questions DROP CONSTRAINT IF EXISTS cloned_questions_assignment_type_check;
ALTER TABLE cloned_questions ADD CONSTRAINT cloned_questions_assignment_type_check
    CHECK (assignment_type = ANY (ARRAY['homework'::text, 'exam'::text, 'lesson_plan_part'::text, 'level_test'::text, 'exercise'::text, 'contest'::text, 'contest_round'::text]));

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_contest_rounds_contest_id ON contest_rounds(contest_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_users_contest_round_id ON contest_round_users(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_users_user_id ON contest_round_users(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_positions_contest_round_id ON contest_round_question_user_positions(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_positions_user_id ON contest_round_question_user_positions(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_positions_question_id ON contest_round_question_user_positions(question_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_multiple_choices_contest_round_id ON contest_round_question_user_multiple_choices(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_multiple_choices_user_id ON contest_round_question_user_multiple_choices(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_multiple_choices_question_id ON contest_round_question_user_multiple_choices(question_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_matchings_contest_round_id ON contest_round_question_user_matchings(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_matchings_user_id ON contest_round_question_user_matchings(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_matchings_question_id ON contest_round_question_user_matchings(question_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_manual_scoring_contest_round_id ON contest_round_question_user_manual_scoring(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_manual_scoring_user_id ON contest_round_question_user_manual_scoring(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_manual_scoring_question_id ON contest_round_question_user_manual_scoring(question_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_labelings_contest_round_id ON contest_round_question_user_labelings(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_labelings_user_id ON contest_round_question_user_labelings(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_labelings_question_id ON contest_round_question_user_labelings(question_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_groups_contest_round_id ON contest_round_question_user_groups(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_groups_user_id ON contest_round_question_user_groups(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_groups_question_id ON contest_round_question_user_groups(question_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_fill_in_blanks_contest_round_id ON contest_round_question_user_fill_in_blanks(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_fill_in_blanks_user_id ON contest_round_question_user_fill_in_blanks(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_question_user_fill_in_blanks_question_id ON contest_round_question_user_fill_in_blanks(question_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_joiner_schools_contest_round_id ON contest_round_joiner_schools(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_joiner_schools_school_id ON contest_round_joiner_schools(school_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_joiner_provinces_contest_round_id ON contest_round_joiner_provinces(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_joiner_provinces_province_id ON contest_round_joiner_provinces(province_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_joiner_persons_contest_round_id ON contest_round_joiner_persons(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_joiner_persons_user_id ON contest_round_joiner_persons(user_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_joiner_classes_contest_round_id ON contest_round_joiner_classes(contest_round_id);
CREATE INDEX IF NOT EXISTS idx_contest_round_joiner_classes_class_id ON contest_round_joiner_classes(class_id);
