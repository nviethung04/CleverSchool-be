-- +migrate Up

CREATE TABLE lessons
(
    id            SERIAL PRIMARY KEY,
    chapter_id    INT          NOT NULL,
    title         VARCHAR(255) NOT NULL,
    description   TEXT,
    status        BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_position INT,
    type          TEXT         NOT NULL DEFAULT 'lesson',
    created_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    FOREIGN KEY (chapter_id) REFERENCES chapters (id) ON DELETE CASCADE
);

CREATE INDEX idx_lessons_chapter ON lessons (chapter_id);
CREATE INDEX idx_lessons_status ON lessons (status);
CREATE INDEX idx_lessons_type ON lessons (type);