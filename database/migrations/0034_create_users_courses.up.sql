-- +migrate Up

CREATE TABLE users_courses
(
    users_id   INT     NOT NULL,
    courses_id INT     NOT NULL,
    start_time DATE,
    end_time   DATE,
    is_current BOOLEAN NOT NULL,
    PRIMARY KEY (users_id, courses_id),
    FOREIGN KEY (users_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (courses_id) REFERENCES courses (id) ON DELETE CASCADE
);

CREATE INDEX idx_users_courses_user ON users_courses (users_id);
CREATE INDEX idx_users_courses_course ON users_courses (courses_id);