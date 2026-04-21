DELETE FROM permissions WHERE permission IN (
  'faculties.index',
  'faculties.show', 
  'faculties.store',
  'faculties.update',
  'faculties.destroy',
  'faculties.restore'
);
