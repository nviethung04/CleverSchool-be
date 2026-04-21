package models

type ProgramRefSubject struct {
	ID        int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ProgramId int64 `gorm:"not null" json:"program_id"`
	SubjectId int64 `gorm:"not null" json:"subject_id"`
}
