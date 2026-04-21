INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh sách bài kiểm tra đánh giá', 'assessments.index', 29, 'Bài kiểm tra đánh giá', TRUE),
  ('Xem chi tiết bài kiểm tra đánh giá', 'assessments.show', 29, 'Bài kiểm tra đánh giá', TRUE),
  ('Tạo bài kiểm tra đánh giá', 'assessments.store', 29, 'Bài kiểm tra đánh giá', TRUE),
  ('Sửa bài kiểm tra đánh giá', 'assessments.update', 29, 'Bài kiểm tra đánh giá', TRUE),
  ('Xóa bài kiểm tra đánh giá', 'assessments.destroy', 29, 'Bài kiểm tra đánh giá', TRUE);
