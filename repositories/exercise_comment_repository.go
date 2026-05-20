package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
)

type ExerciseCommentRepository interface {
    CreateExerciseComment(comment *models.ExerciseComment) error
}

type exerciseCommentRepository struct{}

func NewExerciseCommentRepository() ExerciseCommentRepository { return &exerciseCommentRepository{} }

func (r *exerciseCommentRepository) CreateExerciseComment(comment *models.ExerciseComment) error {
    var existing models.ExerciseComment
    err := db.MasterDB.Where("exercise_id = ? AND student_id = ? AND teacher_id = ?", comment.ExerciseID, comment.StudentID, comment.TeacherID).First(&existing).Error
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



