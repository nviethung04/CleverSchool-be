DELETE FROM permissions
WHERE permission IN (
  'assessment-criteria.index',
  'assessment-criteria.show',
  'assessment-criteria.store',
  'assessment-criteria.update',
  'assessment-criteria.destroy'
);

