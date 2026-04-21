-- Remove contest permissions from all roles
DELETE FROM role_permissions 
WHERE permission_id IN (
  SELECT id FROM permissions 
  WHERE permission IN (
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
