DELETE FROM permissions
WHERE permission IN (
  'assessments.index',
  'assessments.show',
  'assessments.store',
  'assessments.update',
  'assessments.destroy'
);
