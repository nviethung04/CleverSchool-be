DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE permission LIKE 'feedbacks.%' OR permission LIKE 'h5p-contents.%'
);

DELETE FROM permissions
WHERE permission LIKE 'feedbacks.%' OR permission LIKE 'h5p-contents.%';
