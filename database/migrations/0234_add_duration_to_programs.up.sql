ALTER TABLE programs ADD COLUMN duration INTEGER;

UPDATE programs p
SET duration = c.duration
FROM courses c
WHERE c.program_id = p.id;
