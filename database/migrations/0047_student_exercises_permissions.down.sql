DELETE FROM role_permissions rp
USING permissions p
WHERE rp.permission_id = p.id
  AND rp.role_id = 3
  AND p.permission LIKE 'exercises.%';
