package models

type QuestionRefAttribute struct {
	QuestionID        int64   `gorm:"primaryKey"`
	AttributeID       int64   `gorm:"primaryKey;column:question_attribute_id"`
	ParentAttributeID *int64  `gorm:"column:parent_attribute_id"`
	Weight            float64 `gorm:"column:weight;default:0"`

	Question        Question          `gorm:"foreignKey:QuestionID"`
	Attribute       QuestionAttribute `gorm:"foreignKey:AttributeID"`
	ParentAttribute QuestionAttribute `gorm:"foreignKey:ParentAttributeID"`
}
