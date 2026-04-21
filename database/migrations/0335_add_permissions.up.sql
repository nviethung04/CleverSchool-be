INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh sách học kỳ', 'semesters.index', 34, 'Học kỳ', TRUE),
  ('Xem chi tiết học kỳ', 'semesters.show', 34, 'Học kỳ', TRUE),
  ('Tạo học kỳ', 'semesters.store', 34, 'Học kỳ', TRUE),
  ('Sửa học kỳ', 'semesters.update', 34, 'Học kỳ', TRUE),
  ('Xóa học kỳ', 'semesters.destroy', 34, 'Học kỳ', TRUE),
  ('Xem danh sách ngày nghỉ', 'holidays.index', 35, 'Ngày nghỉ', TRUE),
  ('Xem chi tiết ngày nghỉ', 'holidays.show', 35, 'Ngày nghỉ', TRUE),
  ('Tạo ngày nghỉ', 'holidays.store', 35, 'Ngày nghỉ', TRUE),
  ('Sửa ngày nghỉ', 'holidays.update', 35, 'Ngày nghỉ', TRUE),
  ('Xóa ngày nghỉ', 'holidays.destroy', 35, 'Ngày nghỉ', TRUE);

INSERT INTO role_permissions (permission_id, role_id)
SELECT p.id, r.role_id
FROM permissions p
CROSS JOIN (VALUES (1), (2), (3)) AS r(role_id)
WHERE p.permission LIKE 'semesters.%'
   OR p.permission LIKE 'holidays.%';
