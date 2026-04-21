CREATE TABLE topics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    parent_id INT REFERENCES topics(id) ON DELETE SET NULL,
    description TEXT,
    image_url TEXT,
    status BOOLEAN DEFAULT TRUE,

    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);
