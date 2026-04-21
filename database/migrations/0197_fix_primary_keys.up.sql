-- Add primary key constraints to existing tables if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'courses_pkey') THEN
        ALTER TABLE courses ADD CONSTRAINT courses_pkey PRIMARY KEY (id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'users_pkey') THEN
        ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'medias_pkey') THEN
        ALTER TABLE medias ADD CONSTRAINT medias_pkey PRIMARY KEY (id);
    END IF;
END $$;