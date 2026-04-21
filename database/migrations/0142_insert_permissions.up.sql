INSERT INTO permissions (name, permission, sort_position, "group", is_display)
VALUES
  ('Import trường học', 'schools.import', 4, 'Trường học', TRUE),
  ('Export trường học', 'schools.export', 4, 'Trường học', TRUE),
  ('Import lớp học', 'classes.import', 5, 'Lớp học', TRUE),
  ('Export lớp học', 'classes.export', 5, 'Lớp học', TRUE);
