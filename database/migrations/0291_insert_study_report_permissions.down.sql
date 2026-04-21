DELETE FROM permissions
WHERE permission IN (
  'study-reports.index',
  'study-reports.show',
  'study-reports.store',
  'study-reports.update',
  'study-reports.destroy',
  'study-report-criterias.index',
  'study-report-criterias.show',
  'study-report-criterias.store',
  'study-report-criterias.update',
  'study-report-criterias.destroy'
);
