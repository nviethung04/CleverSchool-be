CREATE TABLE question_attributes (
    id SERIAL PRIMARY KEY,
    parent_id BIGINT,
    subject_id BIGINT,
    name VARCHAR(100) NOT NULL,
    level INT,
    weight DECIMAL(5,2) DEFAULT 0,
    description TEXT
);
