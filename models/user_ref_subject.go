package models

type UserRefSubject struct {
	UserID    int64    `gorm:"primaryKey"`
	SubjectID int64    `gorm:"primaryKey"`
	Subject   *Subject `gorm:"foreignKey:SubjectID"`
}
