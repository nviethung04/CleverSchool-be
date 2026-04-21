DROP TABLE IF EXISTS training_levels;

DELETE FROM permissions WHERE permission IN (
  'training-levels.index',
  'training-levels.show',
  'training-levels.store',
  'training-levels.update',
  'training-levels.destroy'
);
