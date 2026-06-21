INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài tập', 'Xem danh sách bài tập', 'exercises.index', 'exercises.index', 18, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'exercises.index');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài tập', 'Tạo bài tập', 'exercises.store', 'exercises.store', 18, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'exercises.store');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài tập', 'Sửa bài tập', 'exercises.update', 'exercises.update', 18, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'exercises.update');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài tập', 'Xóa bài tập', 'exercises.destroy', 'exercises.destroy', 18, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'exercises.destroy');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài tập', 'Xem chi tiết bài tập', 'exercises.show', 'exercises.show', 18, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'exercises.show');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, p.id
FROM permissions p
WHERE p.permission LIKE 'exercises.%'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = 3 AND rp.permission_id = p.id
  );
