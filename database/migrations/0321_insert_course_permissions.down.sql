DELETE FROM permissions
WHERE permission IN (
  'courses.export',
  'courses.import'
);
