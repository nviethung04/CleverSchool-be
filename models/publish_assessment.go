package models

type PublishAssessment struct {
	CourseId     int64 `gorm:"not null" json:"course_id"`
	AssessmentId int64 `gorm:"not null" json:"assessment_id"`
}

func (PublishAssessment) TableName() string {
	return "publish_assessments"
}
