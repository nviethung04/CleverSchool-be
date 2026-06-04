CREATE TABLE lesson_dependencies (
    lesson_id BIGINT NOT NULL,
    dependency_id BIGINT NOT NULL,
    PRIMARY KEY (lesson_id, dependency_id)
);