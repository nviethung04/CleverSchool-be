CREATE TABLE lesson_schedules (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    lesson_id INTEGER NOT NULL,
    shift_id INTEGER,
    scheduled_date DATE NOT NULL,
    week_id INTEGER NOT NULL,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_lesson_schedules_course_id ON lesson_schedules(course_id);
CREATE INDEX idx_lesson_schedules_lesson_id ON lesson_schedules(lesson_id);
CREATE INDEX idx_lesson_schedules_shift_id ON lesson_schedules(shift_id);
CREATE INDEX idx_lesson_schedules_date ON lesson_schedules(scheduled_date);
CREATE INDEX idx_lesson_schedules_week_id ON lesson_schedules(week_id);
