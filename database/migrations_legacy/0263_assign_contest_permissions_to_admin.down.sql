-- Remove contest permissions from admin role
DELETE FROM role_permissions 
WHERE role_id = 1 
AND permission_id IN (
  SELECT p.id 
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
);