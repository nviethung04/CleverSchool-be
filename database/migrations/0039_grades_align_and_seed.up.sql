-- Align grades table with models/grade.go and seed default Khối 1–12 when empty.

ALTER TABLE grades
    ADD COLUMN IF NOT EXISTS number INT,
    ADD COLUMN IF NOT EXISTS name_vn VARCHAR(50),
    ADD COLUMN IF NOT EXISTS name_en VARCHAR(50);

UPDATE grades
SET
    name_vn = COALESCE(NULLIF(name_vn, ''), name),
    name_en = COALESCE(NULLIF(name_en, ''), 'Grade ' || id::text),
    number = COALESCE(number, id)
WHERE deleted_at IS NULL;

INSERT INTO grades (number, name_vn, name_en, name, status, created_at, updated_at)
SELECT
    n,
    'Khối ' || n::text,
    'Grade ' || n::text,
    'Khối ' || n::text,
    TRUE,
    NOW(),
    NOW()
FROM generate_series(1, 12) AS n
WHERE NOT EXISTS (SELECT 1 FROM grades WHERE deleted_at IS NULL);
