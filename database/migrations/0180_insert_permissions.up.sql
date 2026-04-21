INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh sách chương trình', 'programs.index', 22, 'Chương trình', TRUE),
  ('Xem chi tiết chương trình', 'programs.show', 22, 'Chương trình', TRUE),
  ('Tạo chương trình', 'programs.store', 22, 'Chương trình', TRUE),
  ('Sửa chương trình', 'programs.update', 22, 'Chương trình', TRUE),
  ('Xóa chương trình', 'programs.destroy', 22, 'Chương trình', TRUE);
