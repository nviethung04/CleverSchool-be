package repositories

import (
    "be-Clever School/database/db"
    "be-Clever School/models"
)

type HomeworkCommentRepository interface {
    CreateHomeworkComment(comment *models.HomeworkComment) error
}

type homeworkCommentRepository struct{}

func NewHomeworkCommentRepository() HomeworkCommentRepository { return &homeworkCommentRepository{} }

func (r *homeworkCommentRepository) CreateHomeworkComment(comment *models.HomeworkComment) error {
    var existing models.HomeworkComment
    err := db.MasterDB.Where("homework_id = ? AND student_id = ? AND teacher_id = ?", comment.HomeworkID, comment.StudentID, comment.TeacherID).First(&existing).Error
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


