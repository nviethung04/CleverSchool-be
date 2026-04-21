-- +migrate Up

CREATE TABLE users_classes
(
    users_id   INT     NOT NULL,
    classes_id INT     NOT NULL,
    start_time DATE,
    end_time   DATE,
    is_current BOOLEAN NOT NULL,
    PRIMARY KEY (users_id, classes_id),
    FOREIGN KEY (users_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (classes_id) REFERENCES classes (id) ON DELETE CASCADE
);

CREATE INDEX idx_users_classes_user ON users_classes (users_id);
CREATE INDEX idx_users_classes_class ON users_classes (classes_id);