CREATE TABLE semesters (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE holidays (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE semester_ref_holidays (
    semester_id BIGINT NOT NULL REFERENCES semesters(id) ON DELETE CASCADE,
    holiday_id BIGINT NOT NULL REFERENCES holidays(id) ON DELETE CASCADE,
    PRIMARY KEY (semester_id, holiday_id)
);

CREATE TABLE course_ref_semesters (
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    semester_id BIGINT NOT NULL REFERENCES semesters(id) ON DELETE CASCADE,
    PRIMARY KEY (course_id, semester_id)
);

CREATE TABLE study_shifts (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255),
    start_time TIME,
    end_time TIME,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE TABLE course_ref_study_shifts (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    study_shift_id BIGINT NOT NULL REFERENCES study_shifts(id) ON DELETE CASCADE
);

CREATE TABLE weeks (
    id BIGSERIAL PRIMARY KEY,
    year INT NOT NULL,
    week_number INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE lesson_schedules (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    shift_id BIGINT REFERENCES study_shifts(id) ON DELETE SET NULL,
    scheduled_date TIMESTAMP NOT NULL,
    week_id BIGINT REFERENCES weeks(id) ON DELETE SET NULL,
    sort_position INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_lesson_schedules_course_id ON lesson_schedules(course_id);
