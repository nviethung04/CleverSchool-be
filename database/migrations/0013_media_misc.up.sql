CREATE TABLE medias (
    id              BIGSERIAL PRIMARY KEY,
    parent_id       BIGINT,
    folder_id       BIGINT,
    file_name       TEXT NOT NULL DEFAULT '',
    file_path       TEXT NOT NULL DEFAULT '',
    full_path       TEXT NOT NULL DEFAULT '',
    file_type       TEXT,
    file_size       BIGINT,
    file_extension  TEXT,
    disk_name       TEXT,
    static_url      TEXT,
    type            TEXT NOT NULL DEFAULT 'file',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by      BIGINT,
    updated_by      BIGINT,
    deleted_at      TIMESTAMP,
    deleted_by      BIGINT,
    CONSTRAINT chk_medias_type CHECK (type IN ('file', 'folder'))
);

CREATE UNIQUE INDEX idx_medias_file_path_parent_id_type_disk_name
    ON medias (file_path, parent_id, type, disk_name);

CREATE TABLE feedbacks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    title VARCHAR(500),
    content TEXT,
    file_info JSONB,
    role_id BIGINT REFERENCES roles(id) ON DELETE SET NULL,
    status BIGINT,
    response TEXT,
    note TEXT,
    type BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE password_resets (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE history_uses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    course_id BIGINT,
    lesson_id BIGINT,
    activity_ids TEXT,
    average_used NUMERIC(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE lesson_completions (
    id BIGSERIAL PRIMARY KEY,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
