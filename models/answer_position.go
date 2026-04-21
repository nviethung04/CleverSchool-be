package models

type AnswerPosition struct {
	ID         int64  `gorm:"primaryKey"`
	QuestionID *int64 `gorm:"column:question_id"`

	Content   string
	FileInfo  MediaInfo `gorm:"type:jsonb" json:"file_info"`
	IsCorrect bool      `gorm:"default:false"`
	Kind      string    `gorm:"type:KIND_ENUM;not null"`
	Point     float64

	SortPosition    int
	CorrectPosition int
	GroupPosition   int
}
