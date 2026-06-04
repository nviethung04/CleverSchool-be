-- -- Create vocabularies table
-- CREATE TABLE IF NOT EXISTS vocabularies (
--     id BIGSERIAL PRIMARY KEY,
--     word VARCHAR(255) NOT NULL,
--     phonetic VARCHAR(255),
--     translation VARCHAR(500) NOT NULL,
--     definition TEXT,
--     part_of_speech VARCHAR(100),
--     level VARCHAR(50) DEFAULT 'beginner',

--     -- Audio and Image references
--     word_audio_id BIGINT REFERENCES medias(id),
--     image_id BIGINT REFERENCES medias(id),

--     -- Example sentence
--     example_sentence TEXT,
--     example_translation TEXT,
--     example_audio_id BIGINT REFERENCES medias(id),

--     -- Metadata
--     tags VARCHAR(500),
--     difficulty INTEGER DEFAULT 1 CHECK (difficulty >= 1 AND difficulty <= 5),
--     frequency INTEGER DEFAULT 0,

--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     created_by BIGINT,
--     updated_by BIGINT,
--     deleted_at TIMESTAMP,
--     deleted_by BIGINT
-- );

-- -- Create lesson_vocabularies table (nhiều-nhiều relationship)
-- CREATE TABLE IF NOT EXISTS lesson_vocabularies (
--     id BIGSERIAL PRIMARY KEY,
--     lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
--     vocabulary_id BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
--     sort_order INTEGER DEFAULT 0,
--     is_required BOOLEAN DEFAULT TRUE,

--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP,

--     UNIQUE(lesson_id, vocabulary_id)
-- );

-- -- Create student_vocabulary_progress table
-- CREATE TABLE IF NOT EXISTS student_vocabulary_progress (
--     id BIGSERIAL PRIMARY KEY,
--     student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
--     vocabulary_id BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
--     lesson_id BIGINT REFERENCES lessons(id) ON DELETE SET NULL,

--     -- Progress tracking
--     status VARCHAR(50) DEFAULT 'new' CHECK (status IN ('new', 'learning', 'known', 'mastered')),
--     is_known BOOLEAN DEFAULT FALSE,
--     study_count INTEGER DEFAULT 0,
--     correct_count INTEGER DEFAULT 0,
--     last_studied_at TIMESTAMP,
--     mastered_at TIMESTAMP,

--     -- Pronunciation scoring
--     pronunciation_score DECIMAL(5,2),
--     best_pronunciation_score DECIMAL(5,2),
--     pronunciation_attempts INTEGER DEFAULT 0,

--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP,

--     UNIQUE(student_id, vocabulary_id, lesson_id)
-- );

-- -- Create flashcard_sessions table
-- CREATE TABLE IF NOT EXISTS flashcard_sessions (
--     id BIGSERIAL PRIMARY KEY,
--     student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
--     lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,

--     -- Session info
--     total_cards INTEGER DEFAULT 0,
--     completed_cards INTEGER DEFAULT 0,
--     known_cards INTEGER DEFAULT 0,
--     study_duration INTEGER DEFAULT 0,
--     session_score DECIMAL(5,2) DEFAULT 0,

--     started_at TIMESTAMP,
--     completed_at TIMESTAMP,

--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP
-- );

-- -- Create flashcard_activities table
-- CREATE TABLE IF NOT EXISTS flashcard_activities (
--     id BIGSERIAL PRIMARY KEY,
--     session_id BIGINT NOT NULL REFERENCES flashcard_sessions(id) ON DELETE CASCADE,
--     vocabulary_id BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,

--     -- Activity details
--     activity_type VARCHAR(50) NOT NULL CHECK (activity_type IN ('view', 'flip', 'mark_known', 'mark_unknown', 'pronunciation')),
--     response VARCHAR(100),
--     response_time INTEGER DEFAULT 0,
--     pronunciation_score DECIMAL(5,2),

--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- -- Create indexes for better performance
-- CREATE INDEX IF NOT EXISTS idx_vocabularies_word ON vocabularies(word);
-- CREATE INDEX IF NOT EXISTS idx_vocabularies_level ON vocabularies(level);
-- CREATE INDEX IF NOT EXISTS idx_vocabularies_difficulty ON vocabularies(difficulty);
-- CREATE INDEX IF NOT EXISTS idx_vocabularies_deleted_at ON vocabularies(deleted_at);

-- CREATE INDEX IF NOT EXISTS idx_lesson_vocabularies_lesson_id ON lesson_vocabularies(lesson_id);
-- CREATE INDEX IF NOT EXISTS idx_lesson_vocabularies_vocabulary_id ON lesson_vocabularies(vocabulary_id);
-- CREATE INDEX IF NOT EXISTS idx_lesson_vocabularies_sort_order ON lesson_vocabularies(lesson_id, sort_order);

-- CREATE INDEX IF NOT EXISTS idx_student_vocab_progress_student_id ON student_vocabulary_progress(student_id);
-- CREATE INDEX IF NOT EXISTS idx_student_vocab_progress_vocabulary_id ON student_vocabulary_progress(vocabulary_id);
-- CREATE INDEX IF NOT EXISTS idx_student_vocab_progress_lesson_id ON student_vocabulary_progress(lesson_id);
-- CREATE INDEX IF NOT EXISTS idx_student_vocab_progress_status ON student_vocabulary_progress(status);
-- CREATE INDEX IF NOT EXISTS idx_student_vocab_progress_is_known ON student_vocabulary_progress(is_known);

-- CREATE INDEX IF NOT EXISTS idx_flashcard_sessions_student_id ON flashcard_sessions(student_id);
-- CREATE INDEX IF NOT EXISTS idx_flashcard_sessions_lesson_id ON flashcard_sessions(lesson_id);
-- CREATE INDEX IF NOT EXISTS idx_flashcard_sessions_created_at ON flashcard_sessions(created_at);

-- CREATE INDEX IF NOT EXISTS idx_flashcard_activities_session_id ON flashcard_activities(session_id);
-- CREATE INDEX IF NOT EXISTS idx_flashcard_activities_vocabulary_id ON flashcard_activities(vocabulary_id);
-- CREATE INDEX IF NOT EXISTS idx_flashcard_activities_activity_type ON flashcard_activities(activity_type);
-- CREATE INDEX IF NOT EXISTS idx_flashcard_activities_created_at ON flashcard_activities(created_at);

-- -- Add comments for documentation
-- COMMENT ON TABLE vocabularies IS 'Stores vocabulary words with translations, audio, and examples';
-- COMMENT ON TABLE lesson_vocabularies IS 'Links vocabularies to lessons with ordering';
-- COMMENT ON TABLE student_vocabulary_progress IS 'Tracks individual student progress with vocabularies';
-- COMMENT ON TABLE flashcard_sessions IS 'Tracks flashcard study sessions';
-- COMMENT ON TABLE flashcard_activities IS 'Logs individual flashcard interactions';

-- COMMENT ON COLUMN vocabularies.level IS 'Vocabulary difficulty level: beginner, intermediate, advanced';
-- COMMENT ON COLUMN vocabularies.difficulty IS 'Numeric difficulty rating from 1-5';
-- COMMENT ON COLUMN student_vocabulary_progress.status IS 'Learning status: new, learning, known, mastered';
-- COMMENT ON COLUMN flashcard_activities.activity_type IS 'Type of interaction: view, flip, mark_known, mark_unknown, pronunciation';
