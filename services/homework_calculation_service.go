package services

import (
	"be-lms/repositories"
	"fmt"
)

// HomeworkCalculationResult chứa kết quả tính toán
type HomeworkCalculationResult struct {
	Score       float64
	Ratio       float64
	Star        int64
	Exp         float64
	IsCompleted bool
}

type HomeworkCalculationService interface {
	CalculateHomeworkMetrics(homeworkID, userID, lessonID int64) (*HomeworkCalculationResult, error)
}

type homeworkCalculationService struct {
	homeworkUserService      HomeworkUserService
	homeworkUserQuestionRepo repositories.HomeworkUserQuestionRepository
}

func NewHomeworkCalculationService(homeworkUserService HomeworkUserService) HomeworkCalculationService {
	return &homeworkCalculationService{
		homeworkUserService:      homeworkUserService,
		homeworkUserQuestionRepo: repositories.NewHomeworkUserQuestionRepository(),
	}
}

// CalculateHomeworkMetrics tính toán lại exp, star, ratio_score cho homework
func (s *homeworkCalculationService) CalculateHomeworkMetrics(homeworkID, userID, lessonID int64) (*HomeworkCalculationResult, error) {
	result := &HomeworkCalculationResult{}

	// Tính is_completed
	isCompleted, err := s.homeworkUserService.CalculateHomeworkIsCompleted(homeworkID, userID)
	if err != nil {
		return nil, fmt.Errorf("lỗi tính is_completed: %w", err)
	}
	result.IsCompleted = isCompleted

	// Tính tổng điểm từ các bảng homework_question_user_*
	score, err := s.homeworkUserService.CalculateHomeworkScoreService(homeworkID, userID)
	if err != nil {
		return nil, fmt.Errorf("lỗi tính score: %w", err)
	}
	result.Score = score

	// Tính ratio theo logic mới dùng homework_user_questions
	ratio, err := s.homeworkUserService.CalculateHomeworkRatioService(homeworkID, userID)
	if err != nil {
		return nil, fmt.Errorf("lỗi tính ratio: %w", err)
	}
	result.Ratio = ratio

	// Tính exp = tổng (ratio_score * weight) của tất cả các câu hỏi có is_all_correct = true và ratio_score > 0
	exp, err := s.homeworkUserService.CalculateHomeworkExpService(homeworkID, userID)
	if err != nil {
		return nil, fmt.Errorf("lỗi tính exp: %w", err)
	}
	result.Exp = exp

	// Tính star cho bài homework theo lesson hiện tại (tổng star các bản ghi đúng hết)
	_, _, totalStar, err := s.homeworkUserQuestionRepo.GetStarByHomeworkUserLesson(homeworkID, userID, lessonID)
	if err != nil {
		return nil, fmt.Errorf("lỗi tính star: %w", err)
	}
	result.Star = int64(totalStar)

	return result, nil
}

