CREATE TABLE source_questions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE questions (
    id BIGSERIAL PRIMARY KEY,
    source_question_id BIGINT REFERENCES source_questions(id) ON DELETE SET NULL,
    kind kind_enum NOT NULL,
    question_type question_type_enum NOT NULL,
    title VARCHAR(500),
    description TEXT,
    content TEXT,
    keywords TEXT,
    file_info JSONB,
    file_infos JSONB,
    time_limit_seconds INT,
    sort_position INT DEFAULT 0,
    point NUMERIC(10,2) DEFAULT 0,
    is_random SMALLINT DEFAULT 0,
    image_width INT,
    image_height INT,
    status BOOLEAN DEFAULT TRUE,
    display display_enum DEFAULT 'horizontal',
    subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
CREATE INDEX idx_questions_subject_id ON questions(subject_id);
CREATE INDEX idx_questions_type ON questions(question_type);

CREATE TABLE answers (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT REFERENCES questions(id) ON DELETE CASCADE,
    content TEXT,
    file_info JSONB,
    is_correct BOOLEAN DEFAULT FALSE,
    kind kind_enum NOT NULL,
    point NUMERIC(10,2) DEFAULT 0,
    sort_position INT DEFAULT 0
);

CREATE TABLE answer_positions (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    content TEXT,
    file_info JSONB,
    sort_position INT,
    group_position INT DEFAULT 0
);

CREATE TABLE answer_matchings (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    first_item_content TEXT,
    second_item_content TEXT,
    first_item_file_info JSONB,
    second_item_file_info JSONB,
    sort_position INT,
    group_position INT DEFAULT 0,
    matching_group_position INT DEFAULT 0
);

CREATE TABLE answer_coordinates (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    content TEXT,
    x INT, y INT, width INT, height INT,
    group_position INT DEFAULT 0
);

CREATE TABLE answer_groups (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    name VARCHAR(255),
    sort_position INT
);

CREATE TABLE group_answers (
    id BIGSERIAL PRIMARY KEY,
    answer_group_id BIGINT NOT NULL REFERENCES answer_groups(id) ON DELETE CASCADE,
    content TEXT,
    file_info JSONB,
    is_correct BOOLEAN DEFAULT FALSE,
    sort_position INT
);

CREATE TABLE question_attributes (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE question_ref_attributes (
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    question_attribute_id BIGINT NOT NULL REFERENCES question_attributes(id) ON DELETE CASCADE,
    PRIMARY KEY (question_id, question_attribute_id)
);
