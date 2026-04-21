-- +migrate Up

ALTER TABLE programs
ADD COLUMN "detail" JSONB DEFAULT NULL;

