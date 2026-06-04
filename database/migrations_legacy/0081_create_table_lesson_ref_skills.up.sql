CREATE TABLE lesson_ref_skills (
    lesson_id BIGINT NOT NULL,
    skill_id BIGINT NOT NULL,
    PRIMARY KEY (lesson_id, skill_id)
);