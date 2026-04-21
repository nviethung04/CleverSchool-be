-- +migrate Up
-- Add image_info column to contests table
ALTER TABLE contests ADD COLUMN IF NOT EXISTS image_info JSONB;
