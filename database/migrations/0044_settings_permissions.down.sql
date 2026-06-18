DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE permission LIKE 'settings.%');

DELETE FROM permissions WHERE permission LIKE 'settings.%';
