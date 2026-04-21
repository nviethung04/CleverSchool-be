package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	StudyReportTypeStar  = "star"
	StudyReportTypeCheck = "check"
)

type StudyReport struct {
	ID                    int64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                  string   `gorm:"size:255;not null" json:"name"`
	Description           string   `gorm:"type:text" json:"description"`
	SubjectId             int64    `gorm:"not null" json:"subject_id"`
	AssessmentId          int64    `gorm:"not null" json:"assessment_id"`
	AssessmentScoreId     *int64   `json:"assessment_score_id"`
	StudyReportCriteriaId int64    `gorm:"not null" json:"study_report_criteria_id"`
	CourseId              int64    `gorm:"not null" json:"course_id"`
	StudentId             int64    `gorm:"not null" json:"student_id"`
	TeacherId             int64    `gorm:"not null" json:"teacher_id"`
	GeneralComment        string   `gorm:"type:text" json:"general_comment"`
	TotalStar             *int     `json:"total_star"`
	AvgStar               *float64 `gorm:"type:decimal(5,2)" json:"avg_star"`
	IsCompleted           bool     `gorm:"default:false" json:"is_completed"`
	Publish               bool     `gorm:"-" json:"publish"`

	Skills   []StudyReportRefSkillType `gorm:"foreignKey:StudyReportId" json:"skills"`
	Criteria StudyReportCriteria       `gorm:"foreignKey:StudyReportCriteriaId" json:"criteria"`
	Student  User                      `gorm:"foreignKey:StudentId" json:"student"`
	Teacher  User                      `gorm:"foreignKey:TeacherId" json:"teacher"`
	Subject  Subject                   `gorm:"foreignKey:SubjectId" json:"subject"`
	Course   Course                    `gorm:"foreignKey:CourseId" json:"course"`

	// Fix AssignmentScore relation
	AssignmentScore *AssessmentScore `gorm:"references:ID;foreignKey:AssessmentScoreId" json:"assignment_score"`
	Assessment      *Assessment      `gorm:"foreignKey:AssessmentId" json:"assessment"`

	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy int64          `json:"deleted_by"`
}

func (StudyReport) TableName() string {
	return "study_reports"
}

type StudyReportCriteria struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	SubjectId   int64  `gorm:"column:subject_id;not null" json:"subject_id"`
	Name        string `gorm:"column:name;size:255;not null" json:"name"`
	Description string `gorm:"column:description;type:text" json:"description"`
	MaxStar     *int   `gorm:"column:max_star" json:"max_star"`
	Notes       string `gorm:"column:notes;type:jsonb" json:"notes"`

	Skills  []StudyReportSkill `gorm:"foreignKey:CriteriaId" json:"skills"`
	Subject Subject            `gorm:"foreignKey:SubjectId" json:"subject"`

	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	CreatedBy int64          `gorm:"column:created_by" json:"created_by"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy int64          `gorm:"column:updated_by" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy int64          `gorm:"column:deleted_by" json:"deleted_by"`
}

func (StudyReportCriteria) TableName() string {
	return "study_report_criterias"
}

type StudyReportSkill struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	CriteriaId  int64  `gorm:"not null" json:"criteria_id"`
	NameVn      string `gorm:"size:255;not null" json:"name_vn"`
	NameEn      string `gorm:"size:255" json:"name_en"`
	Description string `gorm:"type:text" json:"description"`
	Type        string `gorm:"size:50;not null" json:"type"`

	SortOrder int                    `json:"sort_order"`
	Types     []StudyReportSkillType `gorm:"foreignKey:SkillId" json:"types"`

	Criteria StudyReportCriteria `gorm:"foreignKey:CriteriaId" json:"criteria"`
}

func (StudyReportSkill) TableName() string {
	return "study_report_skills"
}

type StudyReportSkillType struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	SkillId   int64  `gorm:"not null" json:"skill_id"`
	NameVn    string `gorm:"size:255;not null" json:"name_vn"`
	NameEn    string `gorm:"size:255" json:"name_en"`
	ParentId  *int64 `json:"parent_id"`
	Level     int16  `gorm:"default:1" json:"level"`
	SortOrder int    `json:"sort_order"`

	NodeTypes []StudyReportSkillType `gorm:"-" json:"node_types"`
}

func (StudyReportSkillType) TableName() string {
	return "study_report_skill_types"
}

type StudyReportRefSkillType struct {
	ID                    int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	StudyReportId         int64  `gorm:"not null" json:"study_report_id"`
	StudyReportCriteriaId int64  `gorm:"not null" json:"study_report_criteria_id"`
	SkillId               int64  `gorm:"not null" json:"skill_id"`
	SkillTypeId           int64  `gorm:"not null" json:"skill_type_id"`
	Note                  string `gorm:"type:text" json:"note"`
	Star                  *int   `json:"star"`
	IsCheck               bool   `gorm:"default:false" json:"is_check"`

	Skill     StudyReportSkill          `gorm:"foreignKey:SkillId" json:"skill"`
	SkillType StudyReportSkillType      `gorm:"foreignKey:SkillTypeId" json:"skill_type"`
	NodeTypes []StudyReportRefSkillType `gorm:"-" json:"node_types"`
}

func (StudyReportRefSkillType) TableName() string {
	return "study_report_ref_skill_types"
}
