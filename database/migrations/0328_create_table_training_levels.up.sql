CREATE TABLE IF NOT EXISTS training_levels
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    sort_position INTEGER DEFAULT 0,

    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh trình độ đào tạo', 'training-levels.index', 33, 'Trình độ đào tạo', TRUE),
  ('Xem chi tiết trình độ đào tạo', 'training-levels.show', 33, 'Trình độ đào tạo', TRUE),
  ('Tạo trình độ đào tạo', 'training-levels.store', 33, 'Trình độ đào tạo', TRUE),
  ('Sửa trình độ đào tạo', 'training-levels.update', 33, 'Trình độ đào tạo', TRUE),
  ('Xóa trình độ đào tạo', 'training-levels.destroy', 33, 'Trình độ đào tạo', TRUE);
