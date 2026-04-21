-- +migrate Up

CREATE TABLE scorms
(
    id         SERIAL PRIMARY KEY,
    user_id    VARCHAR(255) NOT NULL,
    course_id  VARCHAR(255) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    value      TEXT         NOT NULL,
    type       TEXT         NOT NULL DEFAULT 'scorm',
    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    CONSTRAINT scorms_user_id_course_id_name_key UNIQUE (user_id, course_id, name)
);

CREATE INDEX idx_scorms_user ON scorms (user_id);
CREATE INDEX idx_scorms_course ON scorms (course_id);
CREATE INDEX idx_scorms_type ON scorms (type);