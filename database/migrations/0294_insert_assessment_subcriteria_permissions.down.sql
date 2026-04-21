DELETE FROM permissions
WHERE permission IN (
  'assessment-subcriteria.index',
  'assessment-subcriteria.show',
  'assessment-subcriteria.store',
  'assessment-subcriteria.update',
  'assessment-subcriteria.destroy'
);

