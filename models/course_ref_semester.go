package models

type CourseRefSemester struct {
	CourseId   int64 `gorm:"primaryKey" json:"course_id"`
	SemesterId int64 `gorm:"primaryKey" json:"semester_id"`

	// Relationships
	Course   Course   `gorm:"foreignKey:CourseId"`
	Semester Semester `gorm:"foreignKey:SemesterId"`
}
