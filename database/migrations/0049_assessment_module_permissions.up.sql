INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Bài kiểm tra đánh giá', 'Lưu điểm assessment', 'assessments.save-score', 'assessments.save-score', 21, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'assessments.save-score');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Báo cáo học tập', 'Xem báo cáo học tập', 'study-reports.index', 'study-reports.index', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'study-reports.index');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Báo cáo học tập', 'Tạo báo cáo học tập', 'study-reports.store', 'study-reports.store', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'study-reports.store');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Báo cáo học tập', 'Sửa báo cáo học tập', 'study-reports.update', 'study-reports.update', 22, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'study-reports.update');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Tiêu chí đánh giá', 'Quản lý nhóm tiêu chí', 'assessment-criteria-groups.index', 'assessment-criteria-groups.index', 23, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'assessment-criteria-groups.index');

INSERT INTO permissions ("group", name, description, permission, sort_position, is_display)
SELECT 'Tiêu chí báo cáo', 'Quản lý tiêu chí báo cáo', 'study-report-criterias.index', 'study-report-criterias.index', 24, TRUE
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE permission = 'study-report-criterias.index');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.role_id, p.id
FROM permissions p
CROSS JOIN (VALUES (1), (2)) AS r(role_id)
WHERE p.permission IN (
    'assessments.save-score',
    'study-reports.index', 'study-reports.store', 'study-reports.update',
    'assessment-criteria-groups.index',
    'study-report-criterias.index'
)
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.role_id AND rp.permission_id = p.id
);
