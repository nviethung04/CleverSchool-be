DELETE FROM permissions
WHERE permission IN (
  'faculties.export',
  'faculties.import'
);
