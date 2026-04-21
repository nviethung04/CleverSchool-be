-- Migration Down: Remove response, note, type columns from feedbacks table
-- Version: 0159

ALTER TABLE feedbacks 
DROP COLUMN response,
DROP COLUMN note,
DROP COLUMN type;
