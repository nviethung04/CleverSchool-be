CREATE TABLE headings (
    id SERIAL PRIMARY KEY,
    chapter_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    time INTEGER,
    sort_position INTEGER,
    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

CREATE INDEX idx_headings_chapter_id ON headings(chapter_id);

ALTER TABLE lessons
    ADD COLUMN heading_id BIGINT;

CREATE INDEX idx_lessons_heading_id ON lessons(heading_id);

ALTER TABLE chapters
    DROP COLUMN IF EXISTS time;

ALTER TABLE chapters
    ADD COLUMN time INTEGER NOT NULL DEFAULT 0;
