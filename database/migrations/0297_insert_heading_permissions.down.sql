DELETE FROM permissions
WHERE permission IN (
  'headings.index',
  'headings.show',
  'headings.store',
  'headings.update',
  'headings.destroy'
);
