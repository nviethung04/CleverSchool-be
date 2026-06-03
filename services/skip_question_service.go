package services

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"strconv"
	"time"
)

type SkipQuestionService interface {
	SkipQuestion(req *prot.SkipQuestionRequest, userID int64) (*prot.SkipQuestionResponse, error)
	CheckSubmitHomework(homeworkID, userID int64) (*prot.CheckSubmitHomeworkResponse, error)
}

type skipQuestionService struct {
	homeworkUserRepo             repositories.HomeworkUserRepository
	homeworkUserSkipQuestionRepo repositories.HomeworkUserSkipQuestionRepository
	homeworkRepo                 repositories.HomeworkRepository
	clonedQuestionRepo           repositories.ClonedQuestionRepository
}

func NewSkipQuestionService() SkipQuestionService {
	return &skipQuestionService{
		homeworkUserRepo:             repositories.NewHomeworkUserRepository(),
		homeworkUserSkipQuestionRepo: repositories.NewHomeworkUserSkipQuestionRepository(),
		homeworkRepo:                 repositories.NewHomeworkRepository(),
		clonedQuestionRepo:           repositories.NewClonedQuestionRepository(),
	}
}

func (s *skipQuestionService) SkipQuestion(req *prot.SkipQuestionRequest, userID int64) (*prot.SkipQuestionResponse, error) {
	// Kiểm tra xem đã có record skip question này chưa
	existingSkip, err := s.homeworkUserSkipQuestionRepo.GetByHomeworkUserQuestion(req.HomeworkId, userID, req.QuestionId)
	if err == nil && existingSkip != nil {
		// Nếu đã tồn tại thì không cần thay đổi gì
	} else {
		// Nếu chưa tồn tại, tạo mới
		skipQuestion := &models.HomeworkUserSkipQuestion{
			HomeworkID:   req.HomeworkId,
			UserID:       userID,
			QuestionID:   req.QuestionId,
			DidItAgain:   false,
			DidItAgainAt: nil,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.homeworkUserSkipQuestionRepo.Create(skipQuestion); err != nil {
			return &prot.SkipQuestionResponse{
				Success: false,
				Message: "Failed to create skip question record",
			}, err
		}
	}

	// Cập nhật last_question_id_completed trong bảng homework_users
	if err := s.homeworkUserRepo.UpdateLastQuestionCompleted(req.HomeworkId, userID, req.QuestionId); err != nil {
		return &prot.SkipQuestionResponse{
			Success: false,
			Message: "Failed to update last question completed",
		}, err
	}

	return &prot.SkipQuestionResponse{
		Success:    true,
		Message:    "Question skipped successfully",
		HomeworkId: req.HomeworkId,
		QuestionId: req.QuestionId,
		UserId:     userID,
	}, nil
}

func (s *skipQuestionService) CheckSubmitHomework(homeworkID, userID int64) (*prot.CheckSubmitHomeworkResponse, error) {
	// Bước 1: Lấy danh sách question_id thuộc homework từ bảng cloned_questions
	validQuestionIDs, err := s.getValidQuestionIDs(homeworkID)
	if err != nil {
		return nil, err
	}

	// Bước 2: Xóa dữ liệu trong các bảng homework_question_* với question_id không thuộc homework
	_, err = s.cleanupInvalidQuestionData(homeworkID, userID, validQuestionIDs)
	if err != nil {
		return nil, err
	}

	// Bước 3: Cập nhật question_completed trong bảng homework_users bằng tổng số câu đúng đã trả lời
	err = s.updateQuestionsCompletedWithCorrectAnswers(homeworkID, userID)
	if err != nil {
		return nil, err
	}

	// Bước 4: Lấy danh sách câu hỏi đã skip nhưng chưa làm lại (did_it_again = false)
	skippedNotRedoneIds, err := s.homeworkUserSkipQuestionRepo.ListSkippedNotRedoneQuestionIDs(homeworkID, userID)
	if err != nil {
		return nil, err
	}

	// Bước 5: Lấy danh sách câu hỏi đã skip nhưng đã làm lại (did_it_again = true)
	skippedButRedoneIds, err := s.homeworkUserSkipQuestionRepo.ListSkippedButRedoneQuestionIDs(homeworkID, userID)
	if err != nil {
		return nil, err
	}

	totalQuestions, err := s.homeworkRepo.GetTotalQuestionsFromHomeworks(homeworkID)
	if err != nil {
		return nil, err
	}
	completed, err := s.homeworkUserRepo.GetQuestionsCompleted(homeworkID, userID)
	if err != nil {
		return nil, err
	}
	notCompleted := totalQuestions - completed
	if notCompleted < 0 {
		notCompleted = 0
	}

	return &prot.CheckSubmitHomeworkResponse{
		HomeworkId:                  homeworkID,
		UserId:                      userID,
		SkippedCount:                int64(len(skippedNotRedoneIds)), // Chỉ tính câu chưa làm lại
		SkippedQuestionIds:          skippedNotRedoneIds,             // Chỉ trả về câu chưa làm lại
		NotCompletedCount:           notCompleted,
		SkippedButRedoneCount:       int64(len(skippedButRedoneIds)), // Số lượng câu đã skip nhưng đã làm lại
		SkippedButRedoneQuestionIds: skippedButRedoneIds,             // Danh sách ID câu đã skip nhưng đã làm lại
	}, nil
}

// getValidQuestionIDs lấy danh sách question_id thuộc homework từ bảng cloned_questions
func (s *skipQuestionService) getValidQuestionIDs(homeworkID int64) ([]int64, error) {
	clonedQuestions, err := s.clonedQuestionRepo.GetClonedQuestions(homeworkID, "homework")
	if err != nil {
		return nil, err
	}

	var questionIDs []int64
	for _, question := range clonedQuestions {
		if id, err := strconv.ParseInt(string(question.ID), 10, 64); err == nil {
			questionIDs = append(questionIDs, id)
		}
	}

	return questionIDs, nil
}

// cleanupInvalidQuestionData xóa dữ liệu trong các bảng homework_question_* với question_id không thuộc homework
func (s *skipQuestionService) cleanupInvalidQuestionData(homeworkID, userID int64, validQuestionIDs []int64) (int64, error) {
	if len(validQuestionIDs) == 0 {
		return 0, nil
	}

	var totalDeleted int64

	// Tạo map để lookup nhanh
	validMap := make(map[int64]bool)
	for _, id := range validQuestionIDs {
		validMap[id] = true
	}

	// Xóa trong homework_question_users
	result := db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id NOT IN ?",
		homeworkID, userID, validQuestionIDs).Delete(&models.HomeworkQuestionUser{})
	if result.Error != nil {
		return 0, result.Error
	}
	totalDeleted += result.RowsAffected

	// Xóa trong homework_question_user_positions
	result = db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id NOT IN ?",
		homeworkID, userID, validQuestionIDs).Delete(&models.HomeworkQuestionUserPosition{})
	if result.Error != nil {
		return 0, result.Error
	}
	totalDeleted += result.RowsAffected

	// Xóa trong homework_question_user_matchings
	result = db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id NOT IN ?",
		homeworkID, userID, validQuestionIDs).Delete(&models.HomeworkQuestionUserMatching{})
	if result.Error != nil {
		return 0, result.Error
	}
	totalDeleted += result.RowsAffected

	// Xóa trong homework_question_user_manual_scoring
	result = db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id NOT IN ?",
		homeworkID, userID, validQuestionIDs).Delete(&models.HomeworkQuestionUserManualScoring{})
	if result.Error != nil {
		return 0, result.Error
	}
	totalDeleted += result.RowsAffected

	// Xóa trong homework_question_user_labelings
	result = db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id NOT IN ?",
		homeworkID, userID, validQuestionIDs).Delete(&models.HomeworkQuestionUserLabeling{})
	if result.Error != nil {
		return 0, result.Error
	}
	totalDeleted += result.RowsAffected

	// Xóa trong homework_question_user_groups
	result = db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id NOT IN ?",
		homeworkID, userID, validQuestionIDs).Delete(&models.HomeworkQuestionUserGroup{})
	if result.Error != nil {
		return 0, result.Error
	}
	totalDeleted += result.RowsAffected

	// Xóa trong homework_question_user_fill_in_blanks
	result = db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id NOT IN ?",
		homeworkID, userID, validQuestionIDs).Delete(&models.HomeworkQuestionUserFillInBlank{})
	if result.Error != nil {
		return 0, result.Error
	}
	totalDeleted += result.RowsAffected

	// Xóa trong homework_user_skip_questions
	result = db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id NOT IN ?",
		homeworkID, userID, validQuestionIDs).Delete(&models.HomeworkUserSkipQuestion{})
	if result.Error != nil {
		return 0, result.Error
	}
	totalDeleted += result.RowsAffected

	return totalDeleted, nil
}

// updateQuestionsCompletedWithCorrectAnswers cập nhật question_completed bằng tổng số câu đúng đã trả lời
// Một câu hỏi được tính là đúng khi TẤT CẢ các câu trả lời trong câu hỏi đó đều đúng
func (s *skipQuestionService) updateQuestionsCompletedWithCorrectAnswers(homeworkID, userID int64) error {
	// Lấy danh sách tất cả question_id mà user đã trả lời
	var allAnsweredQuestions []int64

	// Lấy từ tất cả các bảng homework_question_*
	tables := []string{
		"homework_question_users",
		"homework_question_user_positions",
		"homework_question_user_matchings",
		"homework_question_user_manual_scoring",
		"homework_question_user_labelings",
		"homework_question_user_groups",
		"homework_question_user_fill_in_blanks",
	}

	for _, table := range tables {
		var questionIDs []int64
		err := db.MasterDB.Table(table).
			Select("DISTINCT question_id").
			Where("homework_id = ? AND user_id = ?", homeworkID, userID).
			Pluck("question_id", &questionIDs).Error
		if err != nil {
			return err
		}
		allAnsweredQuestions = append(allAnsweredQuestions, questionIDs...)
	}

	// Loại bỏ trùng lặp
	uniqueQuestions := make(map[int64]bool)
	for _, qID := range allAnsweredQuestions {
		uniqueQuestions[qID] = true
	}

	var correctQuestionsCount int64

	// Kiểm tra từng câu hỏi xem có tất cả câu trả lời đều đúng không
	for questionID := range uniqueQuestions {
		isFullyCorrect := true

		// Kiểm tra homework_question_users
		var wrongCount int64
		err := db.MasterDB.Table("homework_question_users").
			Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
				homeworkID, userID, questionID).
			Count(&wrongCount).Error
		if err != nil {
			return err
		}
		if wrongCount > 0 {
			isFullyCorrect = false
		}

		// Kiểm tra homework_question_user_positions
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_positions").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Kiểm tra homework_question_user_matchings
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_matchings").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Kiểm tra homework_question_user_manual_scoring
		//if isFullyCorrect {
		//	err = db.MasterDB.Table("homework_question_user_manual_scoring").
		//		Where("homework_id = ? AND user_id = ? AND question_id = ? AND (is_scored = false OR is_scored IS NULL)",
		//			homeworkID, userID, questionID).
		//		Count(&wrongCount).Error
		//	if err != nil {
		//		return err
		//	}
		//	if wrongCount > 0 {
		//		isFullyCorrect = false
		//	}
		//}

		// Kiểm tra homework_question_user_labelings
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_labelings").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Kiểm tra homework_question_user_groups
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_groups").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Kiểm tra homework_question_user_fill_in_blanks
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_fill_in_blanks").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Nếu tất cả câu trả lời đều đúng, tăng counter
		if isFullyCorrect {
			correctQuestionsCount++
		}
	}

	// Cập nhật question_completed trong bảng homework_users
	return db.MasterDB.Model(&models.HomeworkUser{}).
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Update("questions_completed", correctQuestionsCount).Error
}
