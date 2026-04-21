-- Assign contest permissions to all existing roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE p.permission IN (
  'contests.index',
  'contests.show', 
  'contests.store',
  'contests.update',
  'contests.destroy',
  'contests.restore',
  'contest_rounds.index',
  'contest_rounds.show',
  'contest_rounds.store', 
  'contest_rounds.update',
  'contest_rounds.destroy',
  'contest_rounds.restore'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;
