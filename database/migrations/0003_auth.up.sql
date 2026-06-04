CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT DEFAULT 0,
    name VARCHAR(50) NOT NULL,
    default_page_view VARCHAR(50) NOT NULL DEFAULT 'admin',
    default_page_id BIGINT DEFAULT 0,
    status BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    "group" VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255),
    permission VARCHAR(255) NOT NULL,
    sort_position INT DEFAULT 0,
    is_display BOOLEAN DEFAULT TRUE
);

CREATE TABLE role_permissions (
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    school_id BIGINT REFERENCES schools(id) ON DELETE SET NULL,
    identifier TEXT,
    code TEXT,
    username VARCHAR(100) UNIQUE,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone_number VARCHAR(50),
    address VARCHAR(500),
    status BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    avatar_info JSONB,
    date_of_birth TIMESTAMP,
    parent_id BIGINT,
    last_login_at TIMESTAMP,
    type_teacher SMALLINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);
CREATE INDEX idx_users_school_id ON users(school_id);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

CREATE TABLE user_ref_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    UNIQUE (user_id, role_id)
);
