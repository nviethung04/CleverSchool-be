CREATE TABLE program_ref_subjects (
    id SERIAL PRIMARY KEY,
    program_id INTEGER NOT NULL,
    subject_id INTEGER NOT NULL,
    CONSTRAINT uq_program_subject UNIQUE (program_id, subject_id)
);

CREATE INDEX idx_program_ref_subjects_program_id ON program_ref_subjects(program_id);
CREATE INDEX idx_program_ref_subjects_subject_id ON program_ref_subjects(subject_id);

INSERT INTO program_ref_subjects (program_id, subject_id)
SELECT id AS program_id, subject_id
FROM programs
WHERE subject_id IS NOT NULL
ON CONFLICT (program_id, subject_id) DO NOTHING;
