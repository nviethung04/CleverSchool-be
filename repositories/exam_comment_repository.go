package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
)

type ExamCommentRepository interface {
	CreateExamComment(comment *models.ExamComment) error
}

type examCommentRepository struct{}

func NewExamCommentRepository() ExamCommentRepository {
	return &examCommentRepository{}
}

func (r *examCommentRepository) CreateExamComment(comment *models.ExamComment) error {
	var existing models.ExamComment
	err := db.MasterDB.Where("exam_id = ? AND student_id = ? AND teacher_id = ?", comment.ExamID, comment.StudentID, comment.TeacherID).First(&existing).Error
	if err == nil {
		existing.Content = comment.Content
		existing.UpdatedAt = comment.UpdatedAt
		return db.MasterDB.Save(&existing).Error
	}
	if err != nil && err.Error() != "record not found" {
		return err
	}
	return db.MasterDB.Create(comment).Error
}
