package services

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

type HomeworkUserService interface {
	SaveHomeworkUser(homeworkUser *models.HomeworkUser) error
	CalculateHomeworkScoreService(homeworkID, userID int64) (float64, error)
	SaveHomeworkScoreService(homeworkID, userID int64, score float64) error
	UpdateHomeworkStatusScoringService(homeworkID, userID int64) error
	CalculateHomeworkRatioService(homeworkID, userID int64) (float64, error)
	SaveHomeworkRatioService(homeworkID, userID int64, score float64) error
	CalculateHomeworkExpService(homeworkID, userID int64) (float64, error)
	SaveHomeworkExpService(homeworkID, userID int64, exp float64) error
	SaveHomeworkStarService(homeworkID, userID int64, star int64) error
	CheckHomeworkHasQuestionsNeedManualGrading(homeworkID, userID int64) (bool, error)
	TeacherEvaluate(req *prot.TeacherEvaluateRequest) (*prot.TeacherEvaluateRequest, error)
	CalculateHomeworkIsCompleted(homeworkID, userID int64) (bool, error)
	CheckHomeworkHasManualScoringFromClonedQuestions(homeworkID int64) (bool, error)
	SaveHomeworkHasManualScoringService(homeworkID, userID int64, hasManualScoring bool) error
	CalculateCompletionRate(homeworkID, userID int64) (float64, error)
}

type homeworkUserService struct {
	repo                     repositories.HomeworkUserRepository
	homeworkRepo             repositories.HomeworkRepository
	homeworkUserQuestionRepo repositories.HomeworkUserQuestionRepository
	clonedQuestionService    ClonedQuestionService
	manualScoringRepo        repositories.SaveScoreManualScoringRepository
}

func NewHomeworkUserService(repo repositories.HomeworkUserRepository, homeworkRepo repositories.HomeworkRepository, clonedQuestionService ClonedQuestionService, manualScoringRepo repositories.SaveScoreManualScoringRepository) HomeworkUserService {
	return &homeworkUserService{
		repo:                     repo,
		homeworkRepo:             homeworkRepo,
		homeworkUserQuestionRepo: repositories.NewHomeworkUserQuestionRepository(),
		clonedQuestionService:    clonedQuestionService,
		manualScoringRepo:        manualScoringRepo,
	}
}

func (s *homeworkUserService) SaveHomeworkUser(homeworkUser *models.HomeworkUser) error {
	return s.repo.SaveHomeworkUser(homeworkUser)
}

func (s *homeworkUserService) CalculateHomeworkScoreService(homeworkID, userID int64) (float64, error) {
	total := 0.0
	if v, err := s.repo.SumScoreFillInBlank(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreGroup(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreLabeling(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreManual(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreMatching(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScorePosition(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreUser(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	return total, nil
}

func (s *homeworkUserService) SaveHomeworkScoreService(homeworkID, userID int64, score float64) error {
	return s.repo.UpdateHomeworkUserScore(homeworkID, userID, score)
}

func (s *homeworkUserService) UpdateHomeworkStatusScoringService(homeworkID, userID int64) error {
	return s.repo.UpdateHomeworkUserStatusScoring(homeworkID, userID)
}

func (s *homeworkUserService) CalculateHomeworkRatioService(homeworkID, userID int64) (float64, error) {
	// NEW LOGIC:
	// 1. Query từ bảng cloned_questions lấy tất cả các question (id và weight trung bình, <1 thì =1)
	// 2. So sánh với dữ liệu từ bảng homework_user_questions (chỉ lấy những hàng có is_all_correct = true và ratio_score >= 0)
	// 3. Nếu không có thì điểm của câu hỏi đó = 0
	// 4. Tính tử và mẫu dựa trên đó

	// Bước 1: Lấy tất cả questions từ cloned_questions
	var raw struct {
		Questions json.RawMessage `gorm:"column:questions"`
	}
	err := db.ReplicaDB.Table("cloned_questions").
		Select("questions").
		Where("assignment_id = ? AND assignment_type = ?", homeworkID, "homework").
		First(&raw).Error
	if err != nil {
		return 0, fmt.Errorf("Lỗi lấy cloned_question: %v", err)
	}

	var questions []map[string]interface{}
	if err := json.Unmarshal(raw.Questions, &questions); err != nil {
		return 0, fmt.Errorf("Lỗi parse questions: %v", err)
	}

	// Bước 2: Lấy tất cả records từ homework_user_questions có is_all_correct = true và ratio_score >= 0
	hwqRepo := repositories.NewHomeworkUserQuestionRepository()
	hwqRecords, err := hwqRepo.GetAllCorrectByHomeworkUser(homeworkID, userID)
	if err != nil {
		return 0, fmt.Errorf("Lỗi lấy homework_user_questions: %v", err)
	}

	// Tạo map để tìm nhanh record theo question_id (lấy bản ghi mới nhất)
	hwqMap := make(map[int64]*models.HomeworkUserQuestion)
	for i := range hwqRecords {
		record := &hwqRecords[i]
		existing, ok := hwqMap[record.QuestionID]
		if !ok || record.CreatedAt.After(existing.CreatedAt) || (record.CreatedAt.Equal(existing.CreatedAt) && record.ID > existing.ID) {
			hwqMap[record.QuestionID] = record
		}
	}

	// Bước 3: Tính tử và mẫu
	// Mẫu số chỉ tính các câu có is_all_correct = true và ratio_score > 0 trong homework_user_questions
	numerator := 0.0
	denominator := 0.0

	// Chỉ duyệt qua các câu có trong hwqMap (đã có is_all_correct = true và ratio_score > 0)
	for _, hwqRecord := range hwqMap {
		// Lấy ratio_score và weight từ record
		ratioScore := hwqRecord.RatioScore
		weight := float64(hwqRecord.Weight)
		
		// Nếu weight < 1 thì = 1
		if weight < 1 {
			weight = 1
		}

		// Cộng vào tử và mẫu (chỉ tính các câu có is_all_correct = true và ratio_score > 0)
		numerator += ratioScore * weight
		denominator += weight
	}

	// Bước 4: Tính ratio
	if denominator == 0 {
		return 0, nil
	}
	return numerator / denominator, nil
}

func (s *homeworkUserService) SaveHomeworkRatioService(homeworkID, userID int64, score float64) error {
	return s.repo.UpdateHomeworkUserRatio(homeworkID, userID, score)
}

func (s *homeworkUserService) CheckHomeworkHasQuestionsNeedManualGrading(homeworkID, userID int64) (bool, error) {
	return s.manualScoringRepo.CheckUnscoredQuestionsHomework(homeworkID, userID)
}

func (s *homeworkUserService) TeacherEvaluate(req *prot.TeacherEvaluateRequest) (*prot.TeacherEvaluateRequest, error) {
	rate := req.Rate

	if rate != models.RateExcellent && rate != models.RatePass && rate != models.RateNotAchieved && rate != models.RateNotRatedYet {
		return nil, errors.New("invalid rate")
	}

	req.ObjectType = models.ClonedQuestionTypeHomework

	homeworkUserRepo := repositories.NewHomeworkUserRepository()
	homeworkUser, err := homeworkUserRepo.GetByHomeworkAndUser(req.ObjectId, req.UserId)

	if err != nil {
		return nil, err
	}

	homeworkUser.Rate = rate

	if err := homeworkUserRepo.Update(homeworkUser); err != nil {
		return nil, err
	}

	return req, nil
}

// CalculateHomeworkIsCompleted tính toán xem học sinh đã hoàn thành hết bài homework chưa
// Nếu số lượng câu mà hs đã hoàn thành >= homeworks.total_questions thì là đã hoàn thành
func (s *homeworkUserService) CalculateHomeworkIsCompleted(homeworkID, userID int64) (bool, error) {
	// Lấy total_questions từ homework
	totalQuestions, err := s.homeworkRepo.GetTotalQuestionsFromHomeworks(homeworkID)
	if err != nil {
		return false, fmt.Errorf("Lỗi lấy total_questions: %v", err)
	}

	// Nếu total_questions = 0, coi như chưa hoàn thành
	if totalQuestions == 0 {
		return false, nil
	}

	// Lấy questions_completed từ homework_users
	questionsCompleted, err := s.repo.GetQuestionsCompleted(homeworkID, userID)
	if err != nil {
		return false, fmt.Errorf("Lỗi lấy questions_completed: %v", err)
	}

	// So sánh: nếu questions_completed >= total_questions thì đã hoàn thành
	return questionsCompleted >= totalQuestions, nil
}

// CalculateCompletionRate tính tỷ lệ hoàn thành = (questions_completed / total_questions) * 100, làm tròn lấy phần nguyên
func (s *homeworkUserService) CalculateCompletionRate(homeworkID, userID int64) (float64, error) {
	// Lấy total_questions từ homework
	totalQuestions, err := s.homeworkRepo.GetTotalQuestionsFromHomeworks(homeworkID)
	if err != nil {
		return 0, fmt.Errorf("Lỗi lấy total_questions: %v", err)
	}

	// Nếu total_questions = 0, trả về 0
	if totalQuestions == 0 {
		return 0, nil
	}

	// Lấy questions_completed từ homework_users
	questionsCompleted, err := s.repo.GetQuestionsCompleted(homeworkID, userID)
	if err != nil {
		return 0, fmt.Errorf("Lỗi lấy questions_completed: %v", err)
	}

	// Tính completion_rate = (questions_completed / total_questions) * 100, làm tròn đến số nguyên gần nhất
	rate := float64(questionsCompleted) / float64(totalQuestions) * 100
	return math.Round(rate), nil
}

// CalculateHomeworkExpService tính tổng exp từ các câu hỏi đã làm đúng
// exp = tổng (ratio_score * weight) của tất cả các câu hỏi có is_all_correct = true
func (s *homeworkUserService) CalculateHomeworkExpService(homeworkID, userID int64) (float64, error) {
	// Lấy tất cả records từ homework_user_questions có is_all_correct = true
	hwqRecords, err := s.homeworkUserQuestionRepo.GetAllCorrectByHomeworkUser(homeworkID, userID)
	if err != nil {
		return 0, fmt.Errorf("Lỗi lấy homework_user_questions: %v", err)
	}

	// Tính tổng exp = tổng (ratio_score * weight)
	totalExp := 0.0
	// Tạo map để lấy record mới nhất cho mỗi question_id (tránh duplicate)
	questionMap := make(map[int64]*models.HomeworkUserQuestion)
	for i := range hwqRecords {
		record := &hwqRecords[i]
		existing, ok := questionMap[record.QuestionID]
		if !ok || record.CreatedAt.After(existing.CreatedAt) || (record.CreatedAt.Equal(existing.CreatedAt) && record.ID > existing.ID) {
			questionMap[record.QuestionID] = record
		}
	}

	// Tính tổng exp từ các records mới nhất (chỉ lấy ratio_score > 0)
	for _, record := range questionMap {
		if record.IsAllCorrect && record.RatioScore > 0 {
			exp := record.RatioScore * record.Weight
			totalExp += exp
		}
	}

	return totalExp, nil
}

// SaveHomeworkExpService lưu exp vào homework_users
func (s *homeworkUserService) SaveHomeworkExpService(homeworkID, userID int64, exp float64) error {
	return s.repo.UpdateHomeworkUserExp(homeworkID, userID, exp)
}

// SaveHomeworkStarService lưu star vào homework_users
func (s *homeworkUserService) SaveHomeworkStarService(homeworkID, userID int64, star int64) error {
	return s.repo.UpdateHomeworkUserStar(homeworkID, userID, star)
}

// CheckHomeworkHasManualScoringFromClonedQuestions kiểm tra xem homework có câu hỏi type speaking hoặc writing không
func (s *homeworkUserService) CheckHomeworkHasManualScoringFromClonedQuestions(homeworkID int64) (bool, error) {
	// Lấy questions từ cloned_questions
	var raw struct {
		Questions json.RawMessage `gorm:"column:questions"`
	}
	err := db.ReplicaDB.Table("cloned_questions").
		Select("questions").
		Where("assignment_id = ? AND assignment_type = ?", homeworkID, "homework").
		First(&raw).Error
	if err != nil {
		return false, fmt.Errorf("Lỗi lấy cloned_question: %v", err)
	}

	var questions []map[string]interface{}
	if err := json.Unmarshal(raw.Questions, &questions); err != nil {
		return false, fmt.Errorf("Lỗi parse questions: %v", err)
	}

	// Kiểm tra xem có câu hỏi nào thuộc type speaking hoặc writing không
	for _, q := range questions {
		// Lấy type từ question (có thể là "type" hoặc "question_type")
		var questionType string
		if typeVal, ok := q["type"]; ok {
			if typeStr, ok := typeVal.(string); ok {
				questionType = typeStr
			}
		} else if questionTypeVal, ok := q["question_type"]; ok {
			if typeStr, ok := questionTypeVal.(string); ok {
				questionType = typeStr
			}
		}

		// Kiểm tra nếu là speaking hoặc writing
		if questionType == models.QuestionTypeSpeaking || questionType == models.QuestionTypeWriting {
			return true, nil
		}
	}

	return false, nil
}

// SaveHomeworkHasManualScoringService lưu has_manual_scoring vào homework_users
func (s *homeworkUserService) SaveHomeworkHasManualScoringService(homeworkID, userID int64, hasManualScoring bool) error {
	return s.repo.UpdateHomeworkUserHasManualScoring(homeworkID, userID, hasManualScoring)
}

