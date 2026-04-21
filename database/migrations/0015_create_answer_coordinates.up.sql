-- +migrate Up

CREATE TABLE answer_coordinates
(
    id              BIGSERIAL PRIMARY KEY,
    question_id     INT       NOT NULL,
    content         TEXT,
    file_url        VARCHAR(255),
    kind            kind_enum NOT NULL,
    point           INT,
    position_x      INT,
    position_y      INT,
    position_width  INT,
    position_height INT,
    sort_position   INT,
    type            TEXT      NOT NULL DEFAULT 'answer_coordinate',
    created_at      TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    created_by      INT,
    updated_by      INT,
    FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE CASCADE
);

CREATE INDEX idx_answer_coordinates_question ON answer_coordinates (question_id);
CREATE INDEX idx_answer_coordinates_type ON answer_coordinates (type);