DELETE FROM permissions
WHERE permission IN (
  'programs.index',
  'programs.show',
  'programs.store',
  'programs.update',
  'programs.destroy'
);
