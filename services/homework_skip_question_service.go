package services

import (
	"be-lms/repositories"
)

type HomeworkSkipQuestionService interface {
	UpdateDidItAgainIfSkipped(homeworkID, userID, questionID int64) error
}

type homeworkSkipQuestionService struct {
	homeworkUserSkipQuestionRepo repositories.HomeworkUserSkipQuestionRepository
}

func NewHomeworkSkipQuestionService() HomeworkSkipQuestionService {
	return &homeworkSkipQuestionService{
		homeworkUserSkipQuestionRepo: repositories.NewHomeworkUserSkipQuestionRepository(),
	}
}

func (s *homeworkSkipQuestionService) UpdateDidItAgainIfSkipped(homeworkID, userID, questionID int64) error {
	// Kiểm tra xem có record nào với homework_id, user_id, question_id và did_it_again = false không
	// Nếu có thì cập nhật thành true
	return s.homeworkUserSkipQuestionRepo.UpdateDidItAgain(homeworkID, userID, questionID)
}
