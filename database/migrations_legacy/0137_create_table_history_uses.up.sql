CREATE TABLE history_uses (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(10) NOT NULL, -- 'week' hoặc 'month'
    date DATE NOT NULL,        -- ngày đại diện (bắt buộc)
    usage_time JSONB NOT NULL DEFAULT '{}'::JSONB,
    device JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_type_date UNIQUE (type, date)
);
