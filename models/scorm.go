// models/scorm.go
package models

import "time"

// ScormActivity represents a SCORM learning activity
type ScormActivity struct {
	ID              uint     `gorm:"primaryKey"`
	Title           string   `gorm:"size:255;not null"`
	Description     string   `gorm:"type:text"`
	Version         string   `gorm:"type:varchar(16);not null"` // "1.2" hoặc "2004"
	LaunchURL       string   `gorm:"type:text;not null"`
	MasteryScore    *float64 `gorm:"type:decimal(5,2)"`                 // Điểm đạt yêu cầu
	MaxTimeAllowed  string   `gorm:"type:varchar(32)"`                  // Thời gian tối đa cho phép
	TimeLimitAction string   `gorm:"type:varchar(32)"`                  // Hành động khi hết thời gian
	LaunchData      string   `gorm:"type:text"`                         // Dữ liệu khởi chạy
	DataFromCleverSchool string   `gorm:"type:text"`                         // Dữ liệu từ Clever School
	Status          string   `gorm:"type:varchar(32);default:'active'"` // active/inactive
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TableName specifies the table name for ScormActivity
func (ScormActivity) TableName() string {
	return "scorm_activities"
}

// ScormAttempt represents a user's attempt at a SCORM activity
type ScormAttempt struct {
	ID               string     `gorm:"primaryKey;type:varchar(64)"` // attemptId từ runtime
	ActivityID       uint       `gorm:"index;not null"`
	UserID           uint       `gorm:"index;not null"`
	Version          string     `gorm:"type:varchar(16);not null"`             // SCORM version
	Status           string     `gorm:"type:varchar(32);default:'incomplete'"` // incomplete/completed/passed/failed
	ScoreRaw         *float64   `gorm:"type:decimal(5,2)"`                     // Điểm thô
	ScoreMin         *float64   `gorm:"type:decimal(5,2)"`                     // Điểm tối thiểu
	ScoreMax         *float64   `gorm:"type:decimal(5,2)"`                     // Điểm tối đa
	TotalTime        string     `gorm:"type:varchar(32)"`                      // Tổng thời gian (H:MM:SS hoặc PT0H0M0S)
	SuccessStatus    string     `gorm:"type:varchar(32)"`                      // passed/failed/unknown (SCORM 2004)
	CompletionStatus string     `gorm:"type:varchar(32)"`                      // completed/incomplete/not attempted/unknown (SCORM 2004)
	StartTime        time.Time  `gorm:"not null"`
	EndTime          *time.Time // Thời gian kết thúc
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// TableName specifies the table name for ScormAttempt
func (ScormAttempt) TableName() string {
	return "scorm_attempts"
}

// ScormCMI represents CMI (Computer Managed Instruction) data for SCORM
type ScormCMI struct {
	ID        uint   `gorm:"primaryKey"`
	AttemptID string `gorm:"index;type:varchar(64);not null"`
	Element   string `gorm:"index;type:varchar(255);not null"` // CMI element name
	Value     string `gorm:"type:text"`                        // CMI element value
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for ScormCMI
func (ScormCMI) TableName() string {
	return "scorm_cmi"
}

// ScormSession represents a SCORM session for tracking
type ScormSession struct {
	ID        uint      `gorm:"primaryKey"`
	AttemptID string    `gorm:"index;type:varchar(64);not null"`
	SessionID string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	UserAgent string    `gorm:"type:text"`        // Browser user agent
	IPAddress string    `gorm:"type:varchar(45)"` // IP address
	StartedAt time.Time `gorm:"not null"`
	EndedAt   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for ScormSession
func (ScormSession) TableName() string {
	return "scorm_sessions"
}

