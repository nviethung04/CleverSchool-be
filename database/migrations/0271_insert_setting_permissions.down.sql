DELETE FROM permissions WHERE permission IN (
  'settings.index',
  'settings.show',
  'settings.store',
  'settings.update',
  'settings.destroy'
);
