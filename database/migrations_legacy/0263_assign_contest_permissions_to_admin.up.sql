-- Assign contest permissions to admin role

INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, p.id 
FROM permissions p 
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
AND NOT EXISTS (
  SELECT 1 FROM role_permissions rp 
  WHERE rp.role_id = 1 AND rp.permission_id = p.id
);