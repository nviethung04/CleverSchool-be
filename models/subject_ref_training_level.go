package models

type SubjectRefTrainingLevel struct {
	ID              int64 `gorm:"primaryKey;autoIncrement"`
	SubjectID       int64 `gorm:"primaryKey"`
	TrainingLevelID int64 `gorm:"primaryKey"`
}

func (SubjectRefTrainingLevel) TableName() string {
	return "subject_ref_training_levels"
}
