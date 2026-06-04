-- Migration: Add response, note, type columns to feedbacks table
-- Version: 0159

ALTER TABLE feedbacks 
ADD COLUMN response text COLLATE pg_catalog."default",
ADD COLUMN note text COLLATE pg_catalog."default",
ADD COLUMN type bigint;
