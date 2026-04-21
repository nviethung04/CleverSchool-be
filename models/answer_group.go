package models

type AnswerGroup struct {
	ID         int64       `gorm:"primaryKey"`
	QuestionID *int64      `gorm:"column:question_id"`
	GroupID    *int64      `gorm:"column:group_id"`
	Group      GroupAnswer `gorm:"foreignKey:GroupID"`

	Content  string
	FileInfo MediaInfo `gorm:"type:jsonb" json:"file_info"`
	Kind     string    `gorm:"type:KIND_ENUM;not null"`
	Point    float64

	SortPosition int
}
