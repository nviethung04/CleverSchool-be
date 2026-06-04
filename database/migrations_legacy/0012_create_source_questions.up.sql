-- +migrate Up

CREATE TABLE source_questions
(
    id         SERIAL PRIMARY KEY,
    title      VARCHAR(255) NOT NULL,
    skill      skill_enum   NOT NULL,
    content    TEXT,
    level      level_enum   NOT NULL,
    type       TEXT         NOT NULL DEFAULT 'source_question',
    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT
);

CREATE INDEX idx_source_questions_skill ON source_questions (skill);
CREATE INDEX idx_source_questions_level ON source_questions (level);
CREATE INDEX idx_source_questions_type ON source_questions (type);