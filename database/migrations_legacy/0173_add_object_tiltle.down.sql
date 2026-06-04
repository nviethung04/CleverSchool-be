-- Courses
UPDATE courses
SET name = CONCAT(name, ' (', object_title, ')')
WHERE object_title IS NOT NULL;

ALTER TABLE courses DROP COLUMN object_title;

-- Chapters
UPDATE chapters
SET title = CONCAT(title, ' (', object_title, ')')
WHERE object_title IS NOT NULL;

ALTER TABLE chapters DROP COLUMN object_title;

-- Lessons
UPDATE lessons
SET title = CONCAT(title, ' (', object_title, ')')
WHERE object_title IS NOT NULL;

ALTER TABLE lessons DROP COLUMN object_title;

-- Lesson Plans
UPDATE lesson_plans
SET name = CONCAT(name, ' (', object_title, ')')
WHERE object_title IS NOT NULL;

ALTER TABLE lesson_plans DROP COLUMN object_title;

-- Lesson Plan Parts
UPDATE lesson_plan_parts
SET title = CONCAT(title, ' (', object_title, ')')
WHERE object_title IS NOT NULL;

ALTER TABLE lesson_plan_parts DROP COLUMN object_title;

-- Homeworks
UPDATE homeworks
SET name = CONCAT(name, ' (', object_title, ')')
WHERE object_title IS NOT NULL;

ALTER TABLE homeworks DROP COLUMN object_title;

-- Exams
UPDATE exams
SET name = CONCAT(name, ' (', object_title, ')')
WHERE object_title IS NOT NULL;

ALTER TABLE exams DROP COLUMN object_title;
