-- +migrate Up

CREATE TABLE courses
(
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    status        SMALLINT,
    subject_id    BIGINT,
    school_id     BIGINT,
    sort_position SMALLINT,
    type          VARCHAR(255) NOT NULL DEFAULT 'course',
    created_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE SET NULL,
    FOREIGN KEY (school_id) REFERENCES schools (id) ON DELETE SET NULL
);

CREATE INDEX idx_courses_name ON courses (name);
CREATE INDEX idx_courses_status ON courses (status);
CREATE INDEX idx_courses_subject ON courses (subject_id);
CREATE INDEX idx_courses_type ON courses (type);