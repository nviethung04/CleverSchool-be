DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE permission LIKE 'assessments.%');

DELETE FROM permissions WHERE permission LIKE 'assessments.%';
