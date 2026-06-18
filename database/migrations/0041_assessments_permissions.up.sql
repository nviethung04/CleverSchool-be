INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài kiểm tra đánh giá', 'Xem danh sách bài kiểm tra đánh giá', 'assessments.index', 'assessments.index', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'assessments.index');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài kiểm tra đánh giá', 'Tạo bài kiểm tra đánh giá', 'assessments.store', 'assessments.store', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'assessments.store');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài kiểm tra đánh giá', 'Sửa bài kiểm tra đánh giá', 'assessments.update', 'assessments.update', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'assessments.update');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài kiểm tra đánh giá', 'Xóa bài kiểm tra đánh giá', 'assessments.destroy', 'assessments.destroy', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'assessments.destroy');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài kiểm tra đánh giá', 'Xem chi tiết bài kiểm tra đánh giá', 'assessments.show', 'assessments.show', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'assessments.show');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.role_id, p.id
FROM permissions p
CROSS JOIN (VALUES (1), (2)) AS r(role_id)
WHERE p.permission LIKE 'assessments.%'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.role_id AND rp.permission_id = p.id
  );
