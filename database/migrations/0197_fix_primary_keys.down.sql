-- Remove primary key constraints
ALTER TABLE medias DROP CONSTRAINT IF EXISTS medias_pkey;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE courses DROP CONSTRAINT IF EXISTS courses_pkey;