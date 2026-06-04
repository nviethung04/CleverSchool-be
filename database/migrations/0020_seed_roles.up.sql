-- Seed roles mặc định (khớp models/role.go constants)
INSERT INTO roles (id, name, default_page_view, status, parent_id, default_page_id)
VALUES
    (1, 'Admin', 'admin', TRUE, 0, 0),
    (2, 'Teacher', 'teacher', TRUE, 0, 0),
    (3, 'Student', 'student', TRUE, 0, 0),
    (4, 'School', 'admin', TRUE, 0, 0),
    (5, 'Read Only', 'admin', TRUE, 0, 0)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('roles', 'id'), GREATEST((SELECT MAX(id) FROM roles), 5));

-- Permissions: import đầy đủ từ script riêng hoặc migration bổ sung.
-- Tối thiểu cho internal/dashboard command:
INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'system', 'Internal command', 'Chạy lệnh nội bộ', 'internal.command', 0, FALSE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'internal.command');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, p.id FROM permissions p WHERE p.permission = 'internal.command'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 1 AND rp.permission_id = p.id
  );
