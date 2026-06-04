CREATE TABLE user_positions (
    id SERIAL PRIMARY KEY,
    user_id BIGINT,
    employee_position_id BIGINT,
    note TEXT
);
