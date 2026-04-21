DROP TABLE IF EXISTS course_ref_study_shifts;

CREATE TABLE course_ref_study_shifts (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    shift_id INTEGER NOT NULL,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_course_shift UNIQUE (course_id, shift_id, day_of_week)
);

CREATE INDEX idx_course_ref_study_shifts_course_id ON course_ref_study_shifts(course_id);
CREATE INDEX idx_course_ref_study_shifts_shift_id ON course_ref_study_shifts(shift_id);
CREATE INDEX idx_course_ref_study_shifts_day_of_week ON course_ref_study_shifts(day_of_week);
