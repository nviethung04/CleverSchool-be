-- Đảm bảo HS (role 3) có exercises.index + exercises.show (idempotent; migration 0047 có thể đã chạy).
INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài tập', 'Xem danh sách bài tập', 'exercises.index', 'exercises.index', 18, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'exercises.index');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài tập', 'Xem chi tiết bài tập', 'exercises.show', 'exercises.show', 18, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'exercises.show');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, p.id
FROM permissions p
WHERE p.permission IN ('exercises.index', 'exercises.show')
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = 3 AND rp.permission_id = p.id
);
