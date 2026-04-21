package models

type QuestionAttribute struct {
	ID          int64               `gorm:"primaryKey;autoIncrement"`
	ParentID    *int64              `gorm:"column:parent_id"`
	SubjectId   *int64              `gorm:"column:subject_id"`
	Name        string              `gorm:"size:100;not null"`
	Level       int                 `gorm:"default:0"`
	Weight      float64             `gorm:"default:0"`
	Description string              `gorm:"size:255;"`
	Nodes       []QuestionAttribute `gorm:"foreignKey:ParentID"`
	Status      bool                `gorm:"not null" json:"status"`
}
