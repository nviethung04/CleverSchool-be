CREATE TABLE subject_ref_training_levels (
    id SERIAL PRIMARY KEY,
    subject_id INTEGER NOT NULL,
    training_level_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_subject_ref_training_levels UNIQUE (subject_id, training_level_id)
);

CREATE INDEX idx_subject_ref_training_levels_subject_id ON subject_ref_training_levels(subject_id);
CREATE INDEX idx_subject_ref_training_levels_training_level_id ON subject_ref_training_levels(training_level_id);
