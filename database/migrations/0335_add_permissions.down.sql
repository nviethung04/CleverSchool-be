DELETE FROM permissions
WHERE permission IN (
  'holidays.index',
  'holidays.show',
  'holidays.store',
  'holidays.update',
  'holidays.destroy',
  'semesters.index',
  'semesters.show',
  'semesters.store',
  'semesters.update',
  'semesters.destroy'
);
