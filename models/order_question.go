package models

type QuestionSentence struct {
	ID         int64  `gorm:"primaryKey"`
	QuestionID int64  `json:"question_id"`
	Content    string `json:"content"`
	OrderIndex int    `json:"order_index"` // chỉ số đúng (đáp án)
}
