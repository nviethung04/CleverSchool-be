INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh sách khoa', 'faculties.index', 29, 'Khoa', TRUE),
  ('Xem chi tiết khoa', 'faculties.show', 29, 'Khoa', TRUE),
  ('Tạo khoa', 'faculties.store', 29, 'Khoa', TRUE),
  ('Sửa khoa', 'faculties.update', 29, 'Khoa', TRUE),
  ('Xóa khoa', 'faculties.destroy', 29, 'Khoa', TRUE),
  ('Khôi phục khoa', 'faculties.restore', 29, 'Khoa', TRUE);
