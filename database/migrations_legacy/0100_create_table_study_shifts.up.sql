CREATE TABLE study_shifts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,

    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

CREATE INDEX idx_study_shifts_start_time ON study_shifts(start_time);
