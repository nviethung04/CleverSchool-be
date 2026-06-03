-- Migration Down: Remove response, note, type columns from feedbacks table
-- Version: 0159

ALTER TABLE IF EXISTS feedbacks 
DROP COLUMN IF EXISTS response,
DROP COLUMN IF EXISTS note,
DROP COLUMN IF EXISTS type;
