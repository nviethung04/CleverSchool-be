-- Khôi phục trạng thái migration 0051 cho role School (4).
DELETE FROM role_permissions WHERE role_id = 4;

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
