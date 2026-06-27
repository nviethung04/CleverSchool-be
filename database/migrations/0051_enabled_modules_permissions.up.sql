-- Permissions: feedbacks, h5p-contents; gán role School (4) + bổ sung HS/GV/Admin

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Phản hồi', 'Xem danh sách phản hồi', 'feedbacks.index', 'feedbacks.index', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'feedbacks.index');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Phản hồi', 'Tạo phản hồi', 'feedbacks.store', 'feedbacks.store', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'feedbacks.store');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Phản hồi', 'Sửa phản hồi', 'feedbacks.update', 'feedbacks.update', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'feedbacks.update');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Phản hồi', 'Xoá phản hồi', 'feedbacks.destroy', 'feedbacks.destroy', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'feedbacks.destroy');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Phản hồi', 'Xem chi tiết phản hồi', 'feedbacks.show', 'feedbacks.show', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'feedbacks.show');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'H5P', 'Xem danh sách H5P', 'h5p-contents.index', 'h5p-contents.index', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'h5p-contents.index');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'H5P', 'Tạo H5P', 'h5p-contents.store', 'h5p-contents.store', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'h5p-contents.store');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'H5P', 'Sửa H5P', 'h5p-contents.update', 'h5p-contents.update', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'h5p-contents.update');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'H5P', 'Xoá H5P', 'h5p-contents.destroy', 'h5p-contents.destroy', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'h5p-contents.destroy');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'H5P', 'Xem chi tiết H5P', 'h5p-contents.show', 'h5p-contents.show', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'h5p-contents.show');

-- Admin: đủ feedbacks + h5p
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, p.id
FROM permissions p
WHERE p.permission LIKE 'feedbacks.%' OR p.permission LIKE 'h5p-contents.%'
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 1 AND rp.permission_id = p.id
);

-- Teacher: feedbacks (trừ store), h5p đầy đủ
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, p.id
FROM permissions p
WHERE p.permission IN (
    'feedbacks.index', 'feedbacks.show', 'feedbacks.update', 'feedbacks.destroy',
    'h5p-contents.index', 'h5p-contents.show', 'h5p-contents.store', 'h5p-contents.update', 'h5p-contents.destroy'
)
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 2 AND rp.permission_id = p.id
);

-- Student: feedbacks (trừ destroy)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, p.id
FROM permissions p
WHERE p.permission IN ('feedbacks.index', 'feedbacks.store', 'feedbacks.show', 'feedbacks.update')
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 3 AND rp.permission_id = p.id
);

-- School (4): sao chép quyền teacher, thêm users.index, schools.show/update
INSERT INTO role_permissions (role_id, permission_id)
SELECT 4, rp.permission_id
FROM role_permissions rp
WHERE rp.role_id = 2
AND NOT EXISTS (
    SELECT 1 FROM role_permissions x WHERE x.role_id = 4 AND x.permission_id = rp.permission_id
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 4, p.id
FROM permissions p
WHERE p.permission IN ('users.index', 'schools.show', 'schools.update')
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 4 AND rp.permission_id = p.id
);
