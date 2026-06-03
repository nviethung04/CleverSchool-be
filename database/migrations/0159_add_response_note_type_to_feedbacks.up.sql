-- Migration: Add response, note, type columns to feedbacks table
-- Version: 0159

ALTER TABLE IF EXISTS feedbacks 
ADD COLUMN IF NOT EXISTS response text COLLATE pg_catalog."default",
ADD COLUMN IF NOT EXISTS note text COLLATE pg_catalog."default",
ADD COLUMN IF NOT EXISTS type bigint;
