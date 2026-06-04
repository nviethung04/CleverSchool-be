ALTER TABLE exams ADD COLUMN IF NOT EXISTS assigned_at timestamp without time zone;
ALTER TABLE exams ADD COLUMN IF NOT EXISTS assigned_by bigint;

ALTER TABLE homeworks ADD COLUMN IF NOT EXISTS assigned_at timestamp without time zone;
ALTER TABLE homeworks ADD COLUMN IF NOT EXISTS assigned_by bigint;
