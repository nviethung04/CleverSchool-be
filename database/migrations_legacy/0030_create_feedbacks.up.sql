-- +migrate Up

CREATE TABLE feedbacks
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT,
    content    TEXT,
    type       TEXT NOT NULL DEFAULT 'feedback',
    created_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    created_by INT,
    updated_by INT,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
);

CREATE INDEX idx_feedbacks_user ON feedbacks (user_id);
CREATE INDEX idx_feedbacks_type ON feedbacks (type);