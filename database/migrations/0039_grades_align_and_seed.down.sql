-- Best-effort rollback: remove seeded Khối 1–12 rows only.

DELETE FROM grades
WHERE deleted_at IS NULL
  AND number BETWEEN 1 AND 12
  AND name_vn = 'Khối ' || number::text
  AND name_en = 'Grade ' || number::text;

ALTER TABLE grades
    DROP COLUMN IF EXISTS number,
    DROP COLUMN IF EXISTS name_vn,
    DROP COLUMN IF EXISTS name_en;
