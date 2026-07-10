DELETE FROM role_permissions
WHERE role_id = 3
AND permission_id IN (
    SELECT id FROM permissions WHERE permission IN ('exercises.index', 'exercises.show')
);
