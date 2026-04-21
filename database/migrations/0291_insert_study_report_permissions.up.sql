INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Xem danh sách báo cáo quá trình học', 'study-reports.index', 30, 'Báo cáo quá trình học', TRUE),
  ('Xem chi tiết báo cáo quá trình học', 'study-reports.show', 30, 'Báo cáo quá trình học', TRUE),
  ('Tạo báo cáo quá trình học', 'study-reports.store', 30, 'Báo cáo quá trình học', TRUE),
  ('Sửa báo cáo quá trình học', 'study-reports.update', 30, 'Báo cáo quá trình học', TRUE),
  ('Xóa báo cáo quá trình học', 'study-reports.destroy', 30, 'Báo cáo quá trình học', TRUE),

  ('Xem danh sách tiêu chí báo cáo', 'study-report-criterias.index', 31, 'Tiêu chí báo cáo', TRUE),
  ('Xem chi tiết tiêu chí báo cáo', 'study-report-criterias.show', 31, 'Tiêu chí báo cáo', TRUE),
  ('Tạo tiêu chí báo cáo', 'study-report-criterias.store', 31, 'Tiêu chí báo cáo', TRUE),
  ('Sửa tiêu chí báo cáo', 'study-report-criterias.update', 31, 'Tiêu chí báo cáo', TRUE),
  ('Xóa tiêu chí báo cáo', 'study-report-criterias.destroy', 31, 'Tiêu chí báo cáo', TRUE);
