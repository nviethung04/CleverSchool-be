package models

import (
	"time"

	"gorm.io/gorm"
)

type Lesson struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	ChapterID    int64  `json:"chapter_id"`
	HeadingID    int64  `json:"heading_id"`
	Title        string `json:"title"`
	ObjectTitle  string `gorm:"size:255;not null" json:"object_title"`
	Description  string `json:"description"`
	Status       bool   `json:"status"`
	SortPosition int    `json:"sort_position"`
	Views        int    `json:"views"`

	Chapter      Chapter  `gorm:"foreignKey:ChapterID"`
	Dependencies []Lesson `gorm:"many2many:lesson_dependencies"`
	Skills       []Skill  `gorm:"many2many:lesson_ref_skills"`
	Tags         []Tag    `gorm:"many2many:lesson_ref_tags"`
	Topics       []Topic  `gorm:"many2many:lesson_ref_topics"`

	Author      *User            `gorm:"foreignKey:AuthorId;references:ID" json:"author"`
	Exams       []Exam           `gorm:"many2many:exam_ref_lessons"`
	Homeworks   []Homework       `gorm:"many2many:homework_ref_lessons;joinForeignKey:lesson_id;joinReferences:homework_id" json:"homeworks_default"`
	Exercises   []Exercise       `gorm:"many2many:exercise_ref_lessons"`
	Assessments []Assessment     `gorm:"many2many:assessment_ref_lessons;joinForeignKey:lesson_id;joinReferences:assessment_id"`
	LessonPlans []LessonPlan     `gorm:"many2many:lesson_plan_ref_lessons"`
	Schedules   []LessonSchedule `gorm:"foreignKey:LessonID"`
	Heading     Heading          `gorm:"foreignKey:HeadingID"`

	TeachingPlans []TeachingPlan `gorm:"many2many:lesson_ref_teaching_plans;joinForeignKey:lesson_id;joinReferences:teaching_plan_id"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	AuthorId  int64          `json:"author_id"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
