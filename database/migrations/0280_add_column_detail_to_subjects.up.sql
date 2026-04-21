-- +migrate Up

ALTER TABLE subjects
ADD COLUMN "detail" JSONB DEFAULT NULL;

