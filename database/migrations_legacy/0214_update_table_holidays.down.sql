ALTER TABLE holidays
  ADD COLUMN IF NOT EXISTS semester_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_holidays_semester_id
  ON holidays(semester_id);
