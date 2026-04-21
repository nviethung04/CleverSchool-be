CREATE TABLE user_faculties (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    faculty_id INTEGER NOT NULL,
    CONSTRAINT uq_user_faculties UNIQUE (user_id, faculty_id)
);

CREATE INDEX idx_user_faculties_user_id ON user_faculties(user_id);
CREATE INDEX idx_user_faculties_faculty_id ON user_faculties(faculty_id);
