INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh sách bài tập', 'exercises.index', 23, 'Bài tập', TRUE),
  ('Xem chi tiết bài tập', 'exercises.show', 23, 'Bài tập', TRUE),
  ('Tạo bài tập', 'exercises.store', 23, 'Bài tập', TRUE),
  ('Sửa bài tập', 'exercises.update', 23, 'Bài tập', TRUE),
  ('Xóa bài tập', 'exercises.destroy', 23, 'Bài tập', TRUE);
