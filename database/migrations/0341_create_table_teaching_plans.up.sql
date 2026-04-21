CREATE TABLE IF NOT EXISTS teaching_plans
(
    id          BIGSERIAL PRIMARY KEY,
    title        VARCHAR(255) NOT NULL,
    description TEXT,
    file_info   JSONB,
    file_type   VARCHAR(100),

    status BOOLEAN DEFAULT FALSE,
    approved_by  BIGINT,
    approved_at  timestamp without time zone,
    approved_note TEXT,

    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint
);

INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh giáo án giáo viên', 'teaching-plans.index', 36, 'Giáo án giáo viên', TRUE),
  ('Xem chi tiết giáo án giáo viên', 'teaching-plans.show', 36, 'Giáo án giáo viên', TRUE),
  ('Tạo giáo án giáo viên', 'teaching-plans.store', 36, 'Giáo án giáo viên', TRUE),
  ('Sửa giáo án giáo viên', 'teaching-plans.update', 36, 'Giáo án giáo viên', TRUE),
  ('Xóa giáo án giáo viên', 'teaching-plans.destroy', 36, 'Giáo án giáo viên', TRUE),
  ('Duyệt giáo án giáo viên', 'teaching-plans.approve', 36, 'Giáo án giáo viên', TRUE);
