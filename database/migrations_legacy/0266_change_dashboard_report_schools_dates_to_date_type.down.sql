-- Revert start_date and end_date back to timestamp
ALTER TABLE dashboard_report_schools
ALTER COLUMN start_date TYPE TIMESTAMP USING start_date::TIMESTAMP,
ALTER COLUMN end_date TYPE TIMESTAMP USING end_date::TIMESTAMP;


