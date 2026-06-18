INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Cài đặt hệ thống', 'Xem danh sách cài đặt', 'settings.index', 'settings.index', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'settings.index');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Cài đặt hệ thống', 'Tạo cài đặt', 'settings.store', 'settings.store', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'settings.store');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Cài đặt hệ thống', 'Sửa cài đặt', 'settings.update', 'settings.update', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'settings.update');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Cài đặt hệ thống', 'Xóa cài đặt', 'settings.destroy', 'settings.destroy', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'settings.destroy');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Cài đặt hệ thống', 'Xem chi tiết cài đặt', 'settings.show', 'settings.show', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'settings.show');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, p.id FROM permissions p WHERE p.permission LIKE 'settings.%'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 1 AND rp.permission_id = p.id
  );
