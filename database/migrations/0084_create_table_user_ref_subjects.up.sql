CREATE TABLE user_ref_subjects (
    user_id BIGINT NOT NULL,
    subject_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, subject_id)
);
