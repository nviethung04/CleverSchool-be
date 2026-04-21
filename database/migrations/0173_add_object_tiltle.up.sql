-- Courses
ALTER TABLE courses ADD COLUMN object_title VARCHAR(255);

-- Chapters
ALTER TABLE chapters ADD COLUMN object_title VARCHAR(255);

-- Lessons
ALTER TABLE lessons ADD COLUMN object_title VARCHAR(255);

-- Lesson Plans
ALTER TABLE lesson_plans ADD COLUMN object_title VARCHAR(255);

-- Lesson Plan Parts
ALTER TABLE lesson_plan_parts ADD COLUMN object_title VARCHAR(255);

-- Homeworks (có lesson_id)
ALTER TABLE homeworks ADD COLUMN object_title VARCHAR(255);

-- Exams (có lesson_id)
ALTER TABLE exams ADD COLUMN object_title VARCHAR(255);

-- Courses
UPDATE courses
SET object_title = TRIM(BOTH '()' FROM substring(name FROM '\(([^)]+)\)')),
    name = regexp_replace(name, '\s*\([^)]*\)', '', 'g')
WHERE name ~ '\(.*\)';

-- Chapters
UPDATE chapters
SET object_title = TRIM(BOTH '()' FROM substring(title FROM '\(([^)]+)\)')),
    title = regexp_replace(title, '\s*\([^)]*\)', '', 'g')
WHERE title ~ '\(.*\)';

-- Lessons
UPDATE lessons
SET object_title = TRIM(BOTH '()' FROM substring(title FROM '\(([^)]+)\)')),
    title = regexp_replace(title, '\s*\([^)]*\)', '', 'g')
WHERE title ~ '\(.*\)';

-- Lesson Plans
UPDATE lesson_plans
SET object_title = TRIM(BOTH '()' FROM substring(name FROM '\(([^)]+)\)')),
    name = regexp_replace(name, '\s*\([^)]*\)', '', 'g')
WHERE name ~ '\(.*\)';

-- Lesson Plan Parts
UPDATE lesson_plan_parts
SET object_title = TRIM(BOTH '()' FROM substring(title FROM '\(([^)]+)\)')),
    title = regexp_replace(title, '\s*\([^)]*\)', '', 'g')
WHERE title ~ '\(.*\)';

-- Homeworks
UPDATE homeworks
SET object_title = TRIM(BOTH '()' FROM substring(name FROM '\(([^)]+)\)')),
    name = regexp_replace(name, '\s*\([^)]*\)', '', 'g')
WHERE name ~ '\(.*\)';

-- Exams
UPDATE exams
SET object_title = TRIM(BOTH '()' FROM substring(name FROM '\(([^)]+)\)')),
    name = regexp_replace(name, '\s*\([^)]*\)', '', 'g')
WHERE name ~ '\(.*\)';
