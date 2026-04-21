package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Subject struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	FacultyId   int64  `gorm:"null" json:"faculty_id"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text;not null" json:"description"`
	Status      bool   `gorm:"not null" json:"status"`

	Detail         SubjectDetail   `gorm:"type:jsonb" json:"detail"`
	Faculty        Faculty         `gorm:"foreignKey:FacultyId"`
	TrainingLevels []TrainingLevel `gorm:"many2many:subject_ref_training_levels;joinForeignKey:SubjectId;JoinReferences:TrainingLevelId" json:"training_levels"`

	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy int64      `gorm:"null" json:"created_by"`
	UpdatedBy int64      `gorm:"null" json:"updated_by"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	DeletedBy int64      `gorm:"null" json:"deleted_by,omitempty"`
}

// Add column for VTG
type SubjectDetail struct {
	VTGData SubjectVTGData `json:"vtg_data"`
}

type SubjectVTGData struct {
	Code       string `json:"code"`
	Level      string `json:"level"`
	Enrollment string `json:"enrollment"`
	Time       string `json:"time"`
	Avatar     string `json:"avatar"`
}

func (t SubjectDetail) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *SubjectDetail) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan SubjectDetail: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, t)
}

func (t SubjectVTGData) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *SubjectVTGData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan SubjectVTGData: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, t)
}
