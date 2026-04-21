package dto

import "time"

// VocabularyDTO represents vocabulary data for API responses
type VocabularyDTO struct {
	ID           int64  `json:"id"`
	Word         string `json:"word"`
	Phonetic     string `json:"phonetic"`
	Translation  string `json:"translation"`
	Definition   string `json:"definition"`
	PartOfSpeech string `json:"part_of_speech"`
	Level        string `json:"level"`
	Tags         string `json:"tags"`
	Difficulty   int    `json:"difficulty"`
	Frequency    int    `json:"frequency"`

	// Media files
	WordAudio    *MediaDTO `json:"word_audio"`
	Image        *MediaDTO `json:"image"`
	ExampleAudio *MediaDTO `json:"example_audio"`

	// Example
	ExampleSentence    string `json:"example_sentence"`
	ExampleTranslation string `json:"example_translation"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StudentVocabularyProgressDTO represents student progress with a vocabulary
type StudentVocabularyProgressDTO struct {
	ID           int64 `json:"id"`
	StudentID    int64 `json:"student_id"`
	VocabularyID int64 `json:"vocabulary_id"`
	LessonID     int64 `json:"lesson_id"`

	Status        string     `json:"status"`
	IsKnown       bool       `json:"is_known"`
	StudyCount    int        `json:"study_count"`
	CorrectCount  int        `json:"correct_count"`
	LastStudiedAt *time.Time `json:"last_studied_at"`
	MasteredAt    *time.Time `json:"mastered_at"`

	PronunciationScore     *float64 `json:"pronunciation_score"`
	BestPronunciationScore *float64 `json:"best_pronunciation_score"`
	PronunciationAttempts  int      `json:"pronunciation_attempts"`

	// Vocabulary info included
	Vocabulary *VocabularyDTO `json:"vocabulary,omitempty"`
}

// FlashcardSessionDTO represents a flashcard study session
type FlashcardSessionDTO struct {
	ID             int64      `json:"id"`
	StudentID      int64      `json:"student_id"`
	LessonID       int64      `json:"lesson_id"`
	TotalCards     int        `json:"total_cards"`
	CompletedCards int        `json:"completed_cards"`
	KnownCards     int        `json:"known_cards"`
	StudyDuration  int        `json:"study_duration"`
	SessionScore   float64    `json:"session_score"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `json:"created_at"`

	// Progress percentage
	ProgressPercentage float64 `json:"progress_percentage"`
}

// FlashcardActivityDTO represents an individual flashcard activity
type FlashcardActivityDTO struct {
	ID                 int64     `json:"id"`
	SessionID          int64     `json:"session_id"`
	VocabularyID       int64     `json:"vocabulary_id"`
	ActivityType       string    `json:"activity_type"`
	Response           string    `json:"response"`
	ResponseTime       int       `json:"response_time"`
	PronunciationScore *float64  `json:"pronunciation_score"`
	CreatedAt          time.Time `json:"created_at"`
}

// LessonVocabularyDTO represents vocabulary within a lesson context
type LessonVocabularyDTO struct {
	ID           int64 `json:"id"`
	LessonID     int64 `json:"lesson_id"`
	VocabularyID int64 `json:"vocabulary_id"`
	SortOrder    int   `json:"sort_order"`
	IsRequired   bool  `json:"is_required"`

	Vocabulary *VocabularyDTO                `json:"vocabulary"`
	Progress   *StudentVocabularyProgressDTO `json:"progress,omitempty"` // For student view
}

// FlashcardLessonDTO represents a lesson with its vocabularies for flashcard study
type FlashcardLessonDTO struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`

	// Vocabulary statistics
	TotalVocabularies    int `json:"total_vocabularies"`
	NewVocabularies      int `json:"new_vocabularies"`
	LearningVocabularies int `json:"learning_vocabularies"`
	KnownVocabularies    int `json:"known_vocabularies"`
	MasteredVocabularies int `json:"mastered_vocabularies"`

	// Study session info
	LastSessionAt *time.Time `json:"last_session_at,omitempty"`
	TotalSessions int        `json:"total_sessions"`

	// Vocabularies in lesson
	Vocabularies []LessonVocabularyDTO `json:"vocabularies,omitempty"`
}

// FlashcardStatsDTO represents flashcard statistics for a student
type FlashcardStatsDTO struct {
	StudentID int64 `json:"student_id"`

	// Overall progress
	TotalVocabularies    int `json:"total_vocabularies"`
	NewVocabularies      int `json:"new_vocabularies"`
	LearningVocabularies int `json:"learning_vocabularies"`
	KnownVocabularies    int `json:"known_vocabularies"`
	MasteredVocabularies int `json:"mastered_vocabularies"`

	// Session statistics
	TotalSessions  int     `json:"total_sessions"`
	TotalStudyTime int     `json:"total_study_time"` // in seconds
	AverageScore   float64 `json:"average_score"`
	BestScore      float64 `json:"best_score"`

	// Recent activity
	StudiedToday    int        `json:"studied_today"`
	StudiedThisWeek int        `json:"studied_this_week"`
	LastStudyDate   *time.Time `json:"last_study_date"`
	CurrentStreak   int        `json:"current_streak"` // consecutive days
	LongestStreak   int        `json:"longest_streak"`

	// Pronunciation stats
	AveragePronunciationScore float64 `json:"average_pronunciation_score"`
	BestPronunciationScore    float64 `json:"best_pronunciation_score"`
}
