CREATE TABLE lesson_ref_topics (
    lesson_id BIGINT NOT NULL,
    topic_id BIGINT NOT NULL,
    PRIMARY KEY (lesson_id, topic_id)
);
