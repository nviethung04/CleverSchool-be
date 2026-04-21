package models

type CourseRefSemester struct {
	ID         int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	CourseId   int64 `gorm:"not null" json:"course_id"`
	SemesterId int64 `gorm:"not null" json:"semester_id"`

	// Relationships
	Course   Course   `gorm:"foreignKey:CourseId"`
	Semester Semester `gorm:"foreignKey:SemesterId"`
}
