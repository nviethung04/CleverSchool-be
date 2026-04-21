-- +migrate Up

-- Add Meetings permission for creating meetings
INSERT INTO permissions (name, permission, sort_position, "group", is_display)
SELECT 'Tạo meetings', 'meetings.store', 37, 'Meetings', TRUE
WHERE NOT EXISTS (
  SELECT 1 FROM permissions WHERE permission = 'meetings.store'
);

-- Assign to admin + teacher by default
INSERT INTO role_permissions (permission_id, role_id)
SELECT p.id, r.role_id
FROM permissions p
CROSS JOIN (VALUES (1), (2)) AS r(role_id)
WHERE p.permission = 'meetings.store'
  AND NOT EXISTS (
    SELECT 1
    FROM role_permissions rp
    WHERE rp.permission_id = p.id
      AND rp.role_id = r.role_id
  );
