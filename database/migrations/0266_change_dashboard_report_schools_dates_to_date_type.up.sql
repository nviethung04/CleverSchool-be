-- Change start_date and end_date from timestamp to date
ALTER TABLE dashboard_report_schools
ALTER COLUMN start_date TYPE DATE USING start_date::DATE,
ALTER COLUMN end_date TYPE DATE USING end_date::DATE;


