CREATE TABLE h5p_contents (
    id BIGINT PRIMARY KEY,
    content_id TEXT,
    title TEXT NOT NULL,
    library TEXT,
    parameters JSONB NOT NULL DEFAULT '{}',
    metadata JSONB,
    sort_position INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE h5p_content_scores (
    id BIGSERIAL PRIMARY KEY,
    content_id BIGINT,
    user_id VARCHAR(255) NOT NULL,
    score INT NOT NULL,
    max_score INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    opened TIMESTAMP,
    finished TIMESTAMP,
    "time" TIMESTAMP
);

CREATE TABLE h5p_content_user_data (
    id BIGSERIAL PRIMARY KEY,
    content_id BIGINT,
    user_id VARCHAR(255) NOT NULL,
    sub_content_id INT,
    data_type VARCHAR(255) NOT NULL,
    preload BOOLEAN DEFAULT FALSE,
    invalidate BOOLEAN DEFAULT FALSE,
    context_id VARCHAR(255),
    user_state JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (content_id, user_id, sub_content_id, data_type)
);

CREATE TABLE scorm_activities (
    id BIGSERIAL PRIMARY KEY,
    lesson_id BIGINT,
    title VARCHAR(255),
    package_path TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scorm_attempts (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT REFERENCES scorm_activities(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    attempt_number INT DEFAULT 1,
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scorm_cmi (
    id BIGSERIAL PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES scorm_attempts(id) ON DELETE CASCADE,
    element VARCHAR(255) NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scorm_sessions (
    id BIGSERIAL PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES scorm_attempts(id) ON DELETE CASCADE,
    session_time VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
