package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type JSONRaw json.RawMessage

func (j *JSONRaw) Scan(value interface{}) error {
	if value == nil {
		*j = JSONRaw("[]")
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("JSONRaw: wrong type %T", value)
	}
	*j = append((*j)[0:0], bytes...)
	return nil
}

func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte("[]"), nil
	}
	return []byte(j), nil
}

func (AssessmentCriteriaGroup) TableName() string { return "assessment_criteria_groups" }

type AssessmentCriteriaGroup struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	SubjectId int64          `gorm:"column:subject_id" json:"subject_id"`
	HasFile   bool           `gorm:"column:has_file" json:"has_file"`
	Subject   *Subject       `gorm:"foreignKey:SubjectId"`
	Criteria  []AssessmentCriterion `gorm:"foreignKey:GroupId"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy int64          `gorm:"column:deleted_by" json:"deleted_by"`
}

func (AssessmentCriterion) TableName() string { return "assessment_criteria" }

type AssessmentCriterion struct {
	ID             int64                   `gorm:"primaryKey" json:"id"`
	GroupId        int64                   `gorm:"column:group_id" json:"group_id"`
	Name           string                  `json:"name"`
	Description    string                  `json:"description"`
	HasSubcriteria bool                    `gorm:"column:has_subcriteria" json:"has_subcriteria"`
	MaxScore       float64                 `gorm:"column:max_score" json:"max_score"`
	SortOrder      int                     `gorm:"column:sort_order" json:"sort_order"`
	Subcriteria    []AssessmentSubcriterion `gorm:"foreignKey:CriteriaId"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	CreatedBy      int64                   `json:"created_by"`
	UpdatedBy      int64                   `json:"updated_by"`
	DeletedAt      gorm.DeletedAt          `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy      int64                   `gorm:"column:deleted_by" json:"deleted_by"`
}

func (AssessmentSubcriterion) TableName() string { return "assessment_subcriteria" }

type AssessmentSubcriterion struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	CriteriaId  int64          `gorm:"column:criteria_id" json:"criteria_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	MaxScore    float64        `gorm:"column:max_score" json:"max_score"`
	SortOrder   int            `gorm:"column:sort_order" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy   int64          `gorm:"column:deleted_by" json:"deleted_by"`
}

func (StudyReportCriteria) TableName() string { return "study_report_criterias" }

type StudyReportCriteria struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	SubjectId   int64          `gorm:"column:subject_id" json:"subject_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	MaxStar     int            `gorm:"column:max_star" json:"max_star"`
	Notes       JSONRaw        `gorm:"type:jsonb;default:'[]'" json:"notes"`
	StarSkills  JSONRaw        `gorm:"type:jsonb;default:'[]'" json:"star_skills"`
	CheckSkills JSONRaw        `gorm:"type:jsonb;default:'[]'" json:"check_skills"`
	Subject     *Subject       `gorm:"foreignKey:SubjectId"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (AssessmentScore) TableName() string { return "assessment_scores" }

type AssessmentScore struct {
	ID           int64                    `gorm:"primaryKey" json:"id"`
	UserId       int64                    `gorm:"column:user_id" json:"user_id"`
	AssessmentId int64                    `gorm:"column:assessment_id" json:"assessment_id"`
	CourseId     int64                    `gorm:"column:course_id" json:"course_id"`
	TotalScore   float64                  `gorm:"column:total_score" json:"total_score"`
	FileInfos    MediaInfos               `gorm:"type:jsonb" json:"file_infos"`
	IsScored     bool                     `gorm:"column:is_scored" json:"is_scored"`
	Details      []AssessmentScoreDetail  `gorm:"foreignKey:AssessmentScoreId"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	CreatedBy    int64                    `json:"created_by"`
	UpdatedBy    int64                    `json:"updated_by"`
}

func (AssessmentScoreDetail) TableName() string { return "assessment_score_details" }

type AssessmentScoreDetail struct {
	ID                int64   `gorm:"primaryKey" json:"id"`
	AssessmentScoreId int64   `gorm:"column:assessment_score_id" json:"assessment_score_id"`
	CriteriaId        int64   `gorm:"column:criteria_id" json:"criteria_id"`
	SubcriteriaId     *int64  `gorm:"column:subcriteria_id" json:"subcriteria_id"`
	Score             float64 `json:"score"`
}

func (AssessmentPublish) TableName() string { return "assessment_publishes" }

type AssessmentPublish struct {
	CourseId     int64     `gorm:"primaryKey;column:course_id" json:"course_id"`
	AssessmentId int64     `gorm:"primaryKey;column:assessment_id" json:"assessment_id"`
	Publish      bool      `json:"publish"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (StudyReport) TableName() string { return "study_reports" }

type StudyReport struct {
	ID                     int64     `gorm:"primaryKey" json:"id"`
	Name                   string    `json:"name"`
	Description            string    `json:"description"`
	SubjectId              int64     `gorm:"column:subject_id" json:"subject_id"`
	StudyReportCriteriaId  int64     `gorm:"column:study_report_criteria_id" json:"study_report_criteria_id"`
	AssessmentId           int64     `gorm:"column:assessment_id" json:"assessment_id"`
	CourseId               int64     `gorm:"column:course_id" json:"course_id"`
	StudentId              int64     `gorm:"column:student_id" json:"student_id"`
	TeacherId              int64     `gorm:"column:teacher_id" json:"teacher_id"`
	GeneralComment         string    `gorm:"column:general_comment" json:"general_comment"`
	TotalStar              float64   `gorm:"column:total_star" json:"total_star"`
	AvgStar                float64   `gorm:"column:avg_star" json:"avg_star"`
	Type                   string    `gorm:"column:type" json:"type"`
	TypeValue              string    `gorm:"column:type_value" json:"type_value"`
	IsCompleted            bool      `gorm:"column:is_completed" json:"is_completed"`
	MaxStar                int       `gorm:"column:max_star" json:"max_star"`
	StarSkills             JSONRaw   `gorm:"type:jsonb;default:'[]'" json:"star_skills"`
	CheckSkills            JSONRaw   `gorm:"type:jsonb;default:'[]'" json:"check_skills"`
	Assessment             *Assessment `gorm:"foreignKey:AssessmentId"`
	Student                *User       `gorm:"foreignKey:StudentId"`
	Teacher                *User       `gorm:"foreignKey:TeacherId"`
	Course                 *Course     `gorm:"foreignKey:CourseId"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (StudyReportPublish) TableName() string { return "study_report_publishes" }

type StudyReportPublish struct {
	CourseId     int64     `gorm:"primaryKey;column:course_id" json:"course_id"`
	AssessmentId int64     `gorm:"primaryKey;column:assessment_id" json:"assessment_id"`
	Publish      bool      `json:"publish"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}
