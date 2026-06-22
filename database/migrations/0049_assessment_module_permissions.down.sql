DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE permission IN (
        'assessments.save-score',
        'study-reports.index', 'study-reports.store', 'study-reports.update',
        'assessment-criteria-groups.index',
        'study-report-criterias.index'
    )
);

DELETE FROM permissions WHERE permission IN (
    'assessments.save-score',
    'study-reports.index', 'study-reports.store', 'study-reports.update',
    'assessment-criteria-groups.index',
    'study-report-criterias.index'
);
