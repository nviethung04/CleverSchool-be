package models

type MatchingQuestion struct {
	ID         int64 `gorm:"primaryKey"`
	QuestionID int64
	Items      string // lưu JSON
	Matches    string // lưu JSON
}
