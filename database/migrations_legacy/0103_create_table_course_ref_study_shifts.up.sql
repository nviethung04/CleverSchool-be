CREATE TABLE course_ref_study_shifts (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    shift_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_course_shift UNIQUE (course_id, shift_id)
);

CREATE INDEX idx_course_ref_study_shifts_course_id ON course_ref_study_shifts(course_id);
CREATE INDEX idx_course_ref_study_shifts_shift_id ON course_ref_study_shifts(shift_id);
