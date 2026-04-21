-- +migrate Down

DELETE FROM role_permissions
WHERE permission_id IN (
  SELECT id FROM permissions WHERE permission = 'meetings.store'
);

DELETE FROM permissions
WHERE permission = 'meetings.store';
