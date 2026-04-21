package models

type AnswerMatching struct {
	ID         int64  `gorm:"primaryKey"`
	QuestionID *int64 `gorm:"column:question_id"`

	Content          string
	FileInfo         MediaInfo `gorm:"type:jsonb" json:"file_info"`
	MatchingContent  string
	MatchingFileInfo MediaInfo `gorm:"type:jsonb" json:"matching_file_info"`
	IsCorrect        bool      `gorm:"default:false"`
	Kind             string    `gorm:"type:KIND_ENUM;not null"`
	MatchingKind     string    `gorm:"type:KIND_ENUM;not null"`
	Point            float64

	SortPosition          int
	CorrectPosition       int
	GroupPosition         int
	MatchingGroupPosition int
}
