-- +migrate Up

CREATE TABLE chapters
(
    id            SERIAL PRIMARY KEY,
    course_id     INT          NOT NULL,
    title         VARCHAR(255) NOT NULL,
    description   TEXT,
    status        BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_position INT,
    type          TEXT         NOT NULL DEFAULT 'chapter',
    created_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);

CREATE INDEX idx_chapters_course ON chapters (course_id);
CREATE INDEX idx_chapters_status ON chapters (status);
CREATE INDEX idx_chapters_type ON chapters (type);