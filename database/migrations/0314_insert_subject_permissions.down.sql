DELETE FROM permissions
WHERE permission IN (
  'subjects.export',
  'subjects.import'
);
