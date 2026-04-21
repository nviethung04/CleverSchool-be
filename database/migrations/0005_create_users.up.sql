-- +migrate Up

CREATE TABLE users
(
    id                SERIAL PRIMARY KEY,
    role_id           INT          NOT NULL,
    school_id         BIGINT,
    username          VARCHAR(100) NOT NULL UNIQUE,
    name              VARCHAR(100) NOT NULL,
    password          VARCHAR(100) NOT NULL,
    email             VARCHAR(50) UNIQUE,
    phone_number      VARCHAR(15),
    address           VARCHAR(255),
    type              TEXT         NOT NULL DEFAULT 'user',
    email_verified_at TIMESTAMP,
    remember_token    VARCHAR(100),
    status            BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by        INT,
    updated_by        INT,
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    FOREIGN KEY (school_id) REFERENCES schools (id) ON DELETE SET NULL
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (role_id);
CREATE INDEX idx_users_type ON users (type);
CREATE INDEX idx_users_status ON users (status);