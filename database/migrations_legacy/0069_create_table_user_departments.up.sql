CREATE TABLE user_departments (
    id SERIAL PRIMARY KEY,
    user_id BIGINT,
    department_id BIGINT,
    note TEXT
);
