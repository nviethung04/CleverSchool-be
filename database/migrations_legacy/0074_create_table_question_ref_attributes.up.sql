CREATE TABLE question_ref_attributes (
    question_id BIGINT NOT NULL,
    question_attribute_id BIGINT NOT NULL,
    parent_attribute_id BIGINT,
    weight DECIMAL(5,2) DEFAULT 0,
    PRIMARY KEY (question_id, question_attribute_id)
);
