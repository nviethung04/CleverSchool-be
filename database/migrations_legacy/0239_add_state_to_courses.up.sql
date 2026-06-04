CREATE TYPE course_state AS ENUM ('coming', 'active', 'finished');

ALTER TABLE courses
ADD COLUMN state course_state DEFAULT 'coming' NOT NULL;

UPDATE courses
SET state = CASE
  WHEN now() < start_date THEN 'coming'::course_state
  WHEN now() BETWEEN start_date AND end_date THEN 'active'::course_state
  ELSE 'finished'::course_state
END;
