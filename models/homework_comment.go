package models

import "time"

type HomeworkComment struct {
    ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    HomeworkID int64      `gorm:"column:homework_id" json:"homework_id"`
    LessonID   int64      `gorm:"column:lesson_id" json:"lesson_id"`
    StudentID  int64      `gorm:"column:student_id" json:"student_id"`
    TeacherID  int64      `gorm:"column:teacher_id" json:"teacher_id"`
    Content    string     `gorm:"column:content" json:"content"`
    CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
    UpdatedAt  time.Time  `gorm:"column:updated_at" json:"updated_at"`
    UpdatedBy  int64      `gorm:"column:updated_by" json:"updated_by"`
    DeletedAt  *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
    DeletedBy  int64      `gorm:"column:deleted_by" json:"deleted_by"`
}

func (HomeworkComment) TableName() string { return "homework_comments" }


