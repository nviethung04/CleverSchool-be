DELETE FROM permissions
WHERE permission IN (
  'schools.import',
  'schools.export',
  'classes.import',
  'classes.export'
);
