package models

import (
	"time"

	"gorm.io/gorm"
)

// Vocabulary represents a vocabulary word/phrase
type Vocabulary struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	Word         string `gorm:"size:255;not null" json:"word"`           // Từ vựng
	Phonetic     string `gorm:"size:255" json:"phonetic"`                // Phiên âm
	Translation  string `gorm:"size:500;not null" json:"translation"`    // Dịch nghĩa
	Definition   string `gorm:"type:text" json:"definition"`             // Định nghĩa
	PartOfSpeech string `gorm:"size:100" json:"part_of_speech"`          // Loại từ (noun, verb, adj...)
	Level        string `gorm:"size:50;default:'beginner'" json:"level"` // Mức độ: beginner, intermediate, advanced

	// Audio files
	WordAudioID *int64 `json:"word_audio_id"` // ID của file audio phát âm từ
	WordAudio   *Media `gorm:"foreignKey:WordAudioID" json:"word_audio"`

	// Image
	ImageID *int64 `json:"image_id"` // ID của ảnh minh họa
	Image   *Media `gorm:"foreignKey:ImageID" json:"image"`

	// Example sentence
	ExampleSentence    string `gorm:"type:text" json:"example_sentence"`    // Câu ví dụ
	ExampleTranslation string `gorm:"type:text" json:"example_translation"` // Dịch câu ví dụ
	ExampleAudioID     *int64 `json:"example_audio_id"`                     // ID của file audio câu ví dụ
	ExampleAudio       *Media `gorm:"foreignKey:ExampleAudioID" json:"example_audio"`

	// Metadata
	Tags       string `gorm:"size:500" json:"tags"`        // Tags phân loại
	Difficulty int    `gorm:"default:1" json:"difficulty"` // Độ khó 1-5
	Frequency  int    `gorm:"default:0" json:"frequency"`  // Tần suất sử dụng

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	DeletedBy int64          `json:"deleted_by"`
}

// LessonVocabulary represents the relationship between lessons and vocabularies
type LessonVocabulary struct {
	ID           int64 `gorm:"primaryKey" json:"id"`
	LessonID     int64 `gorm:"not null" json:"lesson_id"`
	VocabularyID int64 `gorm:"not null" json:"vocabulary_id"`
	SortOrder    int   `gorm:"default:0" json:"sort_order"`     // Thứ tự trong bài học
	IsRequired   bool  `gorm:"default:true" json:"is_required"` // Bắt buộc học hay không

	Lesson     Lesson     `gorm:"foreignKey:LessonID" json:"lesson"`
	Vocabulary Vocabulary `gorm:"foreignKey:VocabularyID" json:"vocabulary"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// StudentVocabularyProgress tracks student's progress with vocabularies
type StudentVocabularyProgress struct {
	ID           int64 `gorm:"primaryKey" json:"id"`
	StudentID    int64 `gorm:"not null" json:"student_id"`
	VocabularyID int64 `gorm:"not null" json:"vocabulary_id"`
	LessonID     int64 `json:"lesson_id"` // Bài học liên quan

	// Progress tracking
	Status        string     `gorm:"size:50;default:'new'" json:"status"` // new, learning, known, mastered
	IsKnown       bool       `gorm:"default:false" json:"is_known"`       // Học sinh đánh dấu đã biết
	StudyCount    int        `gorm:"default:0" json:"study_count"`        // Số lần học
	CorrectCount  int        `gorm:"default:0" json:"correct_count"`      // Số lần trả lời đúng
	LastStudiedAt *time.Time `json:"last_studied_at"`                     // Lần học cuối
	MasteredAt    *time.Time `json:"mastered_at"`                         // Thời điểm thành thạo

	// Pronunciation scoring
	PronunciationScore     *float64 `json:"pronunciation_score"`                     // Điểm phát âm (0-100)
	BestPronunciationScore *float64 `json:"best_pronunciation_score"`                // Điểm phát âm cao nhất
	PronunciationAttempts  int      `gorm:"default:0" json:"pronunciation_attempts"` // Số lần thử phát âm

	Student    User       `gorm:"foreignKey:StudentID" json:"student"`
	Vocabulary Vocabulary `gorm:"foreignKey:VocabularyID" json:"vocabulary"`
	Lesson     *Lesson    `gorm:"foreignKey:LessonID" json:"lesson"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// FlashcardSession tracks flashcard study sessions
type FlashcardSession struct {
	ID        int64 `gorm:"primaryKey" json:"id"`
	StudentID int64 `gorm:"not null" json:"student_id"`
	LessonID  int64 `gorm:"not null" json:"lesson_id"`

	// Session info
	TotalCards     int     `gorm:"default:0" json:"total_cards"`     // Tổng số thẻ
	CompletedCards int     `gorm:"default:0" json:"completed_cards"` // Số thẻ đã hoàn thành
	KnownCards     int     `gorm:"default:0" json:"known_cards"`     // Số thẻ đánh dấu đã biết
	StudyDuration  int     `gorm:"default:0" json:"study_duration"`  // Thời gian học (giây)
	SessionScore   float64 `gorm:"default:0" json:"session_score"`   // Điểm session

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`

	Student User   `gorm:"foreignKey:StudentID" json:"student"`
	Lesson  Lesson `gorm:"foreignKey:LessonID" json:"lesson"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// FlashcardActivity tracks individual flashcard activities
type FlashcardActivity struct {
	ID           int64 `gorm:"primaryKey" json:"id"`
	SessionID    int64 `gorm:"not null" json:"session_id"`
	VocabularyID int64 `gorm:"not null" json:"vocabulary_id"`

	// Activity details
	ActivityType       string   `gorm:"size:50;not null" json:"activity_type"` // view, flip, mark_known, pronunciation
	Response           string   `gorm:"size:100" json:"response"`              // known/unknown, pronunciation result
	ResponseTime       int      `gorm:"default:0" json:"response_time"`        // Thời gian phản hồi (ms)
	PronunciationScore *float64 `json:"pronunciation_score"`                   // Điểm phát âm nếu có

	Session    FlashcardSession `gorm:"foreignKey:SessionID" json:"session"`
	Vocabulary Vocabulary       `gorm:"foreignKey:VocabularyID" json:"vocabulary"`

	CreatedAt time.Time `json:"created_at"`
}

// Add indexes for better performance
func (Vocabulary) TableName() string {
	return "vocabularies"
}

func (LessonVocabulary) TableName() string {
	return "lesson_vocabularies"
}

func (StudentVocabularyProgress) TableName() string {
	return "student_vocabulary_progress"
}

func (FlashcardSession) TableName() string {
	return "flashcard_sessions"
}

func (FlashcardActivity) TableName() string {
	return "flashcard_activities"
}
