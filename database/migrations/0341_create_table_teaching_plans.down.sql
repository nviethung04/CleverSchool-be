DROP TABLE IF EXISTS teaching_plans;

DELETE FROM permissions WHERE permission IN (
  'teaching-plans.index',
  'teaching-plans.show',
  'teaching-plans.store',
  'teaching-plans.update',
  'teaching-plans.destroy',
  'teaching-plans.approve'
);
