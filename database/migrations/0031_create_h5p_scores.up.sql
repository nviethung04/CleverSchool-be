-- +migrate Up

CREATE TABLE h5p_scores
(
    id            SERIAL PRIMARY KEY,
    user_id       INT,
    lesson_id     INT,
    parent_sub_id VARCHAR(100),
    sub_id        VARCHAR(100),
    min           INT,
    max           INT,
    raw           INT,
    scaled        DOUBLE PRECISION,
    type          VARCHAR(100) NOT NULL DEFAULT 'answered',
    success       SMALLINT     NOT NULL DEFAULT 0,
    duration      DOUBLE PRECISION,
    created_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL,
    FOREIGN KEY (lesson_id) REFERENCES lessons (id) ON DELETE SET NULL
);

CREATE INDEX idx_h5p_scores_user ON h5p_scores (user_id);
CREATE INDEX idx_h5p_scores_lesson ON h5p_scores (lesson_id);
CREATE INDEX idx_h5p_scores_type ON h5p_scores (type);