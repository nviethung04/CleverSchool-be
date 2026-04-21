-- +migrate Up

ALTER TABLE assessments
    ADD COLUMN IF NOT EXISTS has_file boolean DEFAULT false;


