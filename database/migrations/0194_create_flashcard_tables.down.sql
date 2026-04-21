-- Drop flashcard tables in reverse order due to foreign key dependencies

-- Drop indexes first
DROP INDEX IF EXISTS idx_flashcard_activities_created_at;
DROP INDEX IF EXISTS idx_flashcard_activities_activity_type;
DROP INDEX IF EXISTS idx_flashcard_activities_vocabulary_id;
DROP INDEX IF EXISTS idx_flashcard_activities_session_id;

DROP INDEX IF EXISTS idx_flashcard_sessions_created_at;
DROP INDEX IF EXISTS idx_flashcard_sessions_lesson_id;
DROP INDEX IF EXISTS idx_flashcard_sessions_student_id;

DROP INDEX IF EXISTS idx_student_vocab_progress_is_known;
DROP INDEX IF EXISTS idx_student_vocab_progress_status;
DROP INDEX IF EXISTS idx_student_vocab_progress_lesson_id;
DROP INDEX IF EXISTS idx_student_vocab_progress_vocabulary_id;
DROP INDEX IF EXISTS idx_student_vocab_progress_student_id;

DROP INDEX IF EXISTS idx_lesson_vocabularies_sort_order;
DROP INDEX IF EXISTS idx_lesson_vocabularies_vocabulary_id;
DROP INDEX IF EXISTS idx_lesson_vocabularies_lesson_id;

DROP INDEX IF EXISTS idx_vocabularies_deleted_at;
DROP INDEX IF EXISTS idx_vocabularies_difficulty;
DROP INDEX IF EXISTS idx_vocabularies_level;
DROP INDEX IF EXISTS idx_vocabularies_word;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS flashcard_activities;
DROP TABLE IF EXISTS flashcard_sessions;
DROP TABLE IF EXISTS student_vocabulary_progress;
DROP TABLE IF EXISTS lesson_vocabularies;
DROP TABLE IF EXISTS vocabularies;
