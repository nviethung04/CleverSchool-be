package models

type LessonRefSkill struct {
	LessonID int64 `gorm:"primaryKey"`
	SkillID  int64 `gorm:"primaryKey"`
}
