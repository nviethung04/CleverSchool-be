package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	TypeTheory      = "theory"
	TypePractice    = "practice"
	TypeModule      = "module"
	TypeIntegration = "integration"
)

type Program struct {
	ID int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	// SubjectId       int64     `gorm:"null" json:"subject_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text;not null" json:"description"`
	ImageInfo   MediaInfo `gorm:"type:jsonb" json:"image_info"`
	Status      bool      `gorm:"null" json:"status"`
	Target      string    `gorm:"null" json:"target"`
	Duration    int       `json:"duration"`

	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy int64          `gorm:"null" json:"created_by"`
	UpdatedBy int64          `gorm:"null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`

	Chapters []Chapter `gorm:"foreignKey:ProgramId" json:"chapters"`
	Courses  []Course  `gorm:"foreignKey:ProgramId" json:"courses"`
	// Subject     Subject      `gorm:"foreignKey:SubjectId"`
	Subjects []Subject     `gorm:"many2many:program_ref_subjects;joinForeignKey:ProgramId;JoinReferences:SubjectId" json:"subjects"`
	Type     string        `gorm:"column:type;default:theory;not null" json:"type"`
	Detail   ProgramDetail `gorm:"type:jsonb" json:"detail"`

	BookUrl string `gorm:"null" json:"book_url"`
}

// Add column for VTG
type ProgramDetail struct {
	VTGData ProgramVTGData `json:"vtg_data"`
}

type ProgramVTGData struct {
	Code             string `json:"code"`
	ManagementCode   string `json:"management_code"`
	NumberOfCredits  int    `json:"number_of_credits"`
	Time             string `json:"time"`
	TheoreticalTime  string `json:"theoretical_time"`
	DiscussionTime   string `json:"discussion_time"`
	PracticeTime     string `json:"practice_time"`
	TestTime         string `json:"test_time"`
	StudyModule      string `json:"study_module"`
	NumberOfFrequent int    `json:"number_of_frequent"`
	NumberOfEvaluate int    `json:"number_of_evaluate"`
}

func (t ProgramDetail) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *ProgramDetail) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ProgramDetail: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, t)
}

func (t ProgramVTGData) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *ProgramVTGData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ProgramVTGData: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, t)
}
