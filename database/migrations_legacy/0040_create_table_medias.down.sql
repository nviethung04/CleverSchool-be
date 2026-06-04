-- +migrate Down

DROP INDEX IF EXISTS idx_medias_file_path_parent_id_type_disk_name;

DROP TABLE medias;
