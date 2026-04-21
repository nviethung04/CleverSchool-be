CREATE TABLE semester_ref_holidays (
    id SERIAL PRIMARY KEY,
    semester_id INTEGER NOT NULL,
    holiday_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_semester_ref_holidays UNIQUE (semester_id, holiday_id)
);

CREATE INDEX idx_semester_ref_holidays_semester_id ON semester_ref_holidays(semester_id);
CREATE INDEX idx_semester_ref_holidays_holiday_id ON semester_ref_holidays(holiday_id);
