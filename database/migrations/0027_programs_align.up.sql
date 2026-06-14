ALTER TABLE programs
    ADD COLUMN IF NOT EXISTS subject_id BIGINT REFERENCES subjects(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS image_info JSONB,
    ADD COLUMN IF NOT EXISTS target TEXT;

CREATE INDEX IF NOT EXISTS idx_programs_subject_id ON programs(subject_id);
