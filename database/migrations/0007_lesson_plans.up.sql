CREATE TABLE lesson_plans (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    status BOOLEAN DEFAULT TRUE,
    clone_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE lesson_plan_parts (
    id BIGSERIAL PRIMARY KEY,
    lesson_plan_id BIGINT NOT NULL REFERENCES lesson_plans(id) ON DELETE CASCADE,
    title VARCHAR(255),
    object_title VARCHAR(255) DEFAULT '',
    description TEXT,
    sort_position INT DEFAULT 0,
    cover_image_info JSONB,
    clone_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE lesson_plan_ref_lessons (
    id BIGSERIAL PRIMARY KEY,
    lesson_plan_id BIGINT NOT NULL,
    lesson_id BIGINT NOT NULL,
    course_id BIGINT,
    UNIQUE (lesson_plan_id, lesson_id)
);

CREATE TABLE lesson_plans_lessons (
    lesson_plan_id BIGINT NOT NULL REFERENCES lesson_plans(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    PRIMARY KEY (lesson_plan_id, lesson_id)
);

CREATE TABLE lesson_plan_completes (
    id BIGSERIAL PRIMARY KEY,
    lesson_plan_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
