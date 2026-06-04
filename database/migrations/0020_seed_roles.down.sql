DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE permission = 'internal.command');
DELETE FROM permissions WHERE permission = 'internal.command';
DELETE FROM roles WHERE id IN (1, 2, 3, 4, 5);
