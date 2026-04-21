DELETE FROM permissions
WHERE permission IN (
  'internal.command'
);
