package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"errors"
	"gorm.io/gorm"
)

type SaveCorrectHomeworkRepository interface {
	UpsertHomeworkUserOnCorrect(homeworkID, lessonID, userID, questionID int64, questionType string) error
}

type saveCorrectHomeworkRepository struct{}

func NewSaveCorrectHomeworkRepository() SaveCorrectHomeworkRepository {
	return &saveCorrectHomeworkRepository{}
}

func (r *saveCorrectHomeworkRepository) UpsertHomeworkUserOnCorrect(homeworkID, lessonID, userID, questionID int64, questionType string) error {
	//table := ""
	//switch questionType {
	//case "multiple_choice":
	//	table = "homework_question_users"
	//case "position":
	//	table = "homework_question_user_positions"
	//case "matching":
	//	table = "homework_question_user_matchings"
	//case "manual_scoring":
	//	table = "homework_question_user_manual_scoring"
	//case "labeling":
	//	table = "homework_question_user_labelings"
	//case "group":
	//	table = "homework_question_user_groups"
	//case "fill_in_blank":
	//	table = "homework_question_user_fill_in_blanks"
	//default:
	//	return errors.New("invalid question type")
	//}

	//// Kiểm tra nếu tất cả các hàng đều is_correct=true thì bỏ qua
	//var count int64
	//db.MasterDB.Table(table).Where("homework_id = ? AND user_id = ? AND question_id = ?", homeworkID, userID, questionID).Count(&count)
	//if count > 0 {
	//	var incorrectCount int64
	//	db.MasterDB.Table(table).Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false", homeworkID, userID, questionID).Count(&incorrectCount)
	//	if incorrectCount == 0 {
	//		// Tất cả câu trả lời đều đúng, vẫn cần cập nhật homework_users
	//		// Không return nil ở đây
	//	}
	//}

	var hwUser models.HomeworkUser
	err := db.MasterDB.Where("homework_id = ? AND user_id = ?", homeworkID, userID).First(&hwUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hwUser = models.HomeworkUser{
				HomeworkID:              homeworkID,
				LessonID:                lessonID,
				UserID:                  userID,
				LastQuestionIDCompleted: questionID,
				QuestionsCompleted:      1,
			}
			return db.MasterDB.Create(&hwUser).Error
		}
		return err
	}
	if questionID != hwUser.LastQuestionIDCompleted {
		hwUser.LastQuestionIDCompleted = questionID
		hwUser.QuestionsCompleted += 1
		return db.MasterDB.Save(&hwUser).Error
	}
	return nil
}
