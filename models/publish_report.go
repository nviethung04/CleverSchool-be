package models

type PublishReport struct {
	CourseId     int64 `gorm:"not null" json:"course_id"`
	AssessmentId int64 `gorm:"not null" json:"assessment_id"`
}

func (PublishReport) TableName() string {
	return "publish_reports"
}
