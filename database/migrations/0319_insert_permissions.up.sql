INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh sách Nhóm tiêu chí', 'assessment-criteria-groups.index', 32, 'Nhóm tiêu chí', TRUE),
  ('Xem chi tiết Nhóm tiêu chí', 'assessment-criteria-groups.show', 32, 'Nhóm tiêu chí', TRUE),
  ('Tạo Nhóm tiêu chí', 'assessment-criteria-groups.store', 32, 'Nhóm tiêu chí', TRUE),
  ('Sửa Nhóm tiêu chí', 'assessment-criteria-groups.update', 32, 'Nhóm tiêu chí', TRUE),
  ('Xóa Nhóm tiêu chí', 'assessment-criteria-groups.destroy', 32, 'Nhóm tiêu chí', TRUE);
