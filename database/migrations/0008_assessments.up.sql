-- Shared assessment columns pattern: exams, homeworks, exercises

CREATE TABLE exams (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT REFERENCES programs(id) ON DELETE SET NULL,
    name VARCHAR(255),
    object_title VARCHAR(255) NOT NULL DEFAULT '',
    status SMALLINT DEFAULT 0,
    time_limit BIGINT DEFAULT 0,
    max_score NUMERIC(10,2) DEFAULT 0,
    description TEXT,
    cover_image_info JSONB,
    deadline TIMESTAMP,
    total_questions INT DEFAULT 0,
    is_random_question BOOLEAN DEFAULT FALSE,
    question_form question_form_enum NOT NULL DEFAULT 'question',
    file_infos JSONB,
    clone_info JSONB,
    is_assigned BOOLEAN DEFAULT FALSE,
    assigned_at TIMESTAMP,
    assigned_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE homeworks (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT REFERENCES programs(id) ON DELETE SET NULL,
    name VARCHAR(255),
    object_title VARCHAR(255) NOT NULL DEFAULT '',
    status SMALLINT DEFAULT 0,
    max_score NUMERIC(10,2) DEFAULT 0,
    description TEXT,
    cover_image_info JSONB,
    total_questions INT DEFAULT 0,
    is_random_question BOOLEAN DEFAULT FALSE,
    question_form question_form_enum NOT NULL DEFAULT 'question',
    file_infos JSONB,
    clone_info JSONB,
    is_assigned BOOLEAN DEFAULT FALSE,
    assigned_at TIMESTAMP,
    assigned_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE exercises (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT REFERENCES programs(id) ON DELETE SET NULL,
    name VARCHAR(255),
    object_title VARCHAR(255) NOT NULL DEFAULT '',
    status SMALLINT DEFAULT 0,
    time_limit BIGINT DEFAULT 0,
    max_score NUMERIC(10,2) DEFAULT 0,
    description TEXT,
    cover_image_info JSONB,
    deadline TIMESTAMP,
    total_questions INT DEFAULT 0,
    is_random_question BOOLEAN DEFAULT FALSE,
    question_form question_form_enum NOT NULL DEFAULT 'question',
    file_infos JSONB,
    clone_info JSONB,
    is_assigned BOOLEAN DEFAULT FALSE,
    assigned_at TIMESTAMP,
    assigned_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE exam_ref_lessons (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    course_id BIGINT REFERENCES courses(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP,
    assigned_by BIGINT,
    UNIQUE (lesson_id, exam_id)
);

CREATE TABLE homework_ref_lessons (
    id BIGSERIAL PRIMARY KEY,
    homework_id BIGINT NOT NULL REFERENCES homeworks(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    course_id BIGINT REFERENCES courses(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP,
    assigned_by BIGINT,
    UNIQUE (lesson_id, homework_id)
);

CREATE TABLE exercise_ref_lessons (
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    course_id BIGINT REFERENCES courses(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP,
    assigned_by BIGINT,
    UNIQUE (lesson_id, exercise_id)
);

CREATE TABLE cloned_questions (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT,
    assignment_id BIGINT NOT NULL,
    assignment_type VARCHAR(50) NOT NULL,
    questions JSONB NOT NULL DEFAULT '[]',
    clone_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
CREATE INDEX idx_cloned_questions_assignment ON cloned_questions(assignment_id, assignment_type);
