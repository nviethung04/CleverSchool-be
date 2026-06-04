-- +migrate Down

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'kind_enum') THEN
DROP TYPE kind_enum;
END IF;
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'level_enum') THEN
DROP TYPE level_enum;
END IF;
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'question_type_enum') THEN
DROP TYPE question_type_enum;
END IF;
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'skill_enum') THEN
DROP TYPE skill_enum;
END IF;
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'topic_enum') THEN
DROP TYPE topic_enum;
END IF;
END $$;