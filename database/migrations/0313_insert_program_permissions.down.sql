DELETE FROM permissions
WHERE permission IN (
  'programs.export',
  'programs.import'
);
