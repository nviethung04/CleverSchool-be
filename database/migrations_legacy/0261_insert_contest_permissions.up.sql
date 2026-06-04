INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh sách cuộc thi', 'contests.index', 26, 'Cuộc thi', TRUE),
  ('Xem chi tiết cuộc thi', 'contests.show', 26, 'Cuộc thi', TRUE),
  ('Tạo cuộc thi', 'contests.store', 26, 'Cuộc thi', TRUE),
  ('Sửa cuộc thi', 'contests.update', 26, 'Cuộc thi', TRUE),
  ('Xóa cuộc thi', 'contests.destroy', 26, 'Cuộc thi', TRUE),
  ('Khôi phục cuộc thi', 'contests.restore', 26, 'Cuộc thi', TRUE),
  ('Xem danh sách vòng thi', 'contest_rounds.index', 27, 'Vòng thi', TRUE),
  ('Xem chi tiết vòng thi', 'contest_rounds.show', 27, 'Vòng thi', TRUE),
  ('Tạo vòng thi', 'contest_rounds.store', 27, 'Vòng thi', TRUE),
  ('Sửa vòng thi', 'contest_rounds.update', 27, 'Vòng thi', TRUE),
  ('Xóa vòng thi', 'contest_rounds.destroy', 27, 'Vòng thi', TRUE),
  ('Khôi phục vòng thi', 'contest_rounds.restore', 27, 'Vòng thi', TRUE);
