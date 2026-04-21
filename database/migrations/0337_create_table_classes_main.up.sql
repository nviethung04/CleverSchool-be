
CREATE TABLE IF NOT EXISTS classes_main
(
    id bigserial,
    name text COLLATE pg_catalog."default",
    school_id bigint,
    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint,
    CONSTRAINT classes_main_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_classes_main_school_id ON classes_main(school_id);
CREATE INDEX IF NOT EXISTS idx_classes_main_deleted_at ON classes_main(deleted_at);

