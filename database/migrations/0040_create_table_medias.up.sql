CREATE TABLE medias (
    id              BIGSERIAL PRIMARY KEY,
    parent_id       BIGINT,
    folder_id       BIGINT,
    file_name       TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    full_path       TEXT NOT NULL,
    file_type       TEXT,
    file_size       BIGINT,
    file_extension  TEXT,
    disk_name       TEXT,
    static_url      TEXT,
    type            TEXT NOT NULL DEFAULT 'file',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by      INT,
    updated_by      INT,
    CONSTRAINT chk_medias_type CHECK (type IN ('file', 'folder'))
);

CREATE UNIQUE INDEX idx_medias_file_path_parent_id_type_disk_name ON medias (file_path, parent_id, type, disk_name);
