DELETE FROM permissions
WHERE permission IN (
  'exercises.index',
  'exercises.show',
  'exercises.store',
  'exercises.update',
  'exercises.destroy'
);
