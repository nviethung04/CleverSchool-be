package models

func (AnswerCoordinates) TableName() string {
	return "answer_coordinates"
}

type AnswerCoordinates struct {
	ID         int64  `gorm:"primaryKey"`
	QuestionID *int64 `gorm:"column:question_id"`

	Content  string
	Kind     string    `gorm:"type:KIND_ENUM;not null"`
	FileInfo MediaInfo `gorm:"type:jsonb" json:"file_info"`
	Point    float64

	PositionX      int
	PositionY      int
	PositionWidth  int
	PositionHeight int

	SortPosition  int
	GroupPosition int
}
