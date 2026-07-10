-- Role School (4): đồng bộ quyền theo GetSchoolPermissions() — phạm vi trường, không sao chép teacher.
DELETE FROM role_permissions WHERE role_id = 4;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 4, p.id
FROM permissions p
WHERE p.permission IN (
    'users.index', 'users.store', 'users.update', 'users.show',
    'schools.show', 'schools.update',
    'classes.index', 'classes.store', 'classes.update', 'classes.destroy', 'classes.show',
    'courses.index', 'courses.store', 'courses.update', 'courses.show',
    'grades.index', 'grades.show',
    'subjects.index', 'subjects.show'
)
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 4 AND rp.permission_id = p.id
);
