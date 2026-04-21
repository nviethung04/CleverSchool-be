-- +migrate Up

CREATE TABLE answer_groups
(
    id            BIGSERIAL PRIMARY KEY,
    question_id   INT       NOT NULL,
    group_id      INT       NOT NULL,
    content       TEXT,
    file_url      VARCHAR(255),
    kind          kind_enum NOT NULL,
    point         INT,
    sort_position INT,
    type          TEXT      NOT NULL DEFAULT 'answer_group',
    created_at    TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    created_by    INT,
    updated_by    INT,
    FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES group_answers (id) ON DELETE CASCADE
);

CREATE INDEX idx_answer_groups_question ON answer_groups (question_id);
CREATE INDEX idx_answer_groups_group ON answer_groups (group_id);
CREATE INDEX idx_answer_groups_type ON answer_groups (type);