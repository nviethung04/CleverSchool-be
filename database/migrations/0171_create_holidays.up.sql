-- Create holidays table
CREATE TABLE IF NOT EXISTS holidays (
    id BIGSERIAL PRIMARY KEY,
    semester_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    type VARCHAR(100) NOT NULL DEFAULT 'day',
    status BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_holidays_semester_id ON holidays(semester_id);
CREATE INDEX IF NOT EXISTS idx_holidays_start_date ON holidays(start_date);
CREATE INDEX IF NOT EXISTS idx_holidays_end_date ON holidays(end_date);
CREATE INDEX IF NOT EXISTS idx_holidays_type ON holidays(type);
