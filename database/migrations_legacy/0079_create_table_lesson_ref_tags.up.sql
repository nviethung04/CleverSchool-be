CREATE TABLE lesson_ref_tags (
    lesson_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY (lesson_id, tag_id)
);