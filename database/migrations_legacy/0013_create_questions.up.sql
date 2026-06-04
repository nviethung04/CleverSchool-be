-- +migrate Up

CREATE TABLE questions
(
    id                 SERIAL PRIMARY KEY,
    source_question_id INT,
    kind               kind_enum          NOT NULL,
    question_type      question_type_enum NOT NULL,
    skill              skill_enum         NOT NULL,
    level              level_enum         NOT NULL,
    topic              topic_enum         NOT NULL,
    title              VARCHAR(255),
    content            TEXT,
    file_url           VARCHAR(255),
    time_limit_seconds INT,
    sort_position      INT,
    point              INT,
    is_random          SMALLINT           NOT NULL DEFAULT 0,
    image_width        INT,
    image_height       INT,
    type               TEXT               NOT NULL DEFAULT 'question',
    created_at         TIMESTAMP                   DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP                   DEFAULT CURRENT_TIMESTAMP,
    created_by         INT,
    updated_by         INT,
    FOREIGN KEY (source_question_id) REFERENCES source_questions (id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES users (id) ON DELETE SET NULL
);

CREATE INDEX idx_questions_source ON questions (source_question_id);
CREATE INDEX idx_questions_type ON questions (question_type);
CREATE INDEX idx_questions_skill ON questions (skill);
CREATE INDEX idx_questions_level ON questions (level);
CREATE INDEX idx_questions_type_field ON questions (type);