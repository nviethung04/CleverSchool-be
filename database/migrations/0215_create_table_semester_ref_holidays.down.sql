-- Down
DROP INDEX IF EXISTS idx_semester_ref_holidays_semester_id;
DROP INDEX IF EXISTS idx_semester_ref_holidays_holiday_id;

DROP TABLE IF EXISTS semester_ref_holidays;
