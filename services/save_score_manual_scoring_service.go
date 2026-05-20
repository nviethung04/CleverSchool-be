package services

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"fmt"
	"strconv"
)

type ManualScoringService interface {
	SaveExamManualScoringService(req prot.SaveAnswerManualRequest, userID int64) error
	SaveExerciseManualScoringService(req prot.SaveAnswerManualRequest, userID int64) error
	SaveHomeworkManualScoringService(req prot.SaveAnswerManualRequest, userID int64) error
}

type SaveScoreManualScoringService interface {
	SaveScoreManualScoring(req *prot.SaveScoreManualRequest, scoringBy int64) error
	SaveScoreManualScoringExercise(req *prot.SaveScoreManualRequest, scoringBy int64) error
	SaveScoreManualScoringHomework(req *prot.SaveScoreManualRequest, scoringBy int64) error
}

type ManualScoringRepository interface {
	SaveExamManualScoringRepository(scoring *models.ExamQuestionUserManualScoring) error
	SaveExerciseManualScoringRepository(scoring *models.ExerciseQuestionUserManualScoring) error
	SaveHomeworkManualScoringRepository(scoring *models.HomeworkQuestionUserManualScoring) error
}

type manualScoringService struct {
	repo                        repositories.ManualScoringRepository
	correctRepo                 repositories.SaveCorrectHomeworkRepository
	homeworkUserQuestionService HomeworkUserQuestionService
}

type saveScoreManualScoringService struct {
	repo                        repositories.SaveScoreManualScoringRepository
	examUserService             ExamUserService
	homeworkUserService         HomeworkUserService
	homeworkUserQuestionService HomeworkUserQuestionService
	homeworkCalculationService  HomeworkCalculationService
	userStarExpService          UserStarExpService
}

func NewManualScoringService(repo repositories.ManualScoringRepository, homeworkUserQuestionService HomeworkUserQuestionService) ManualScoringService {
	return &manualScoringService{
		repo:                        repo,
		correctRepo:                 repositories.NewSaveCorrectHomeworkRepository(),
		homeworkUserQuestionService: homeworkUserQuestionService,
	}
}

func NewSaveScoreManualScoringService(repo repositories.SaveScoreManualScoringRepository, examUserService ExamUserService, homeworkUserService HomeworkUserService, homeworkUserQuestionService HomeworkUserQuestionService, homeworkCalculationService HomeworkCalculationService, userStarExpService UserStarExpService) SaveScoreManualScoringService {
	return &saveScoreManualScoringService{
		repo:                        repo,
		examUserService:             examUserService,
		homeworkUserService:         homeworkUserService,
		homeworkUserQuestionService: homeworkUserQuestionService,
		homeworkCalculationService:  homeworkCalculationService,
		userStarExpService:          userStarExpService,
	}
}

// SaveExamManualScoringService SaveExamManualScoring lưu câu trả lời cần chấm điểm thủ công cho bài thi
func (s *manualScoringService) SaveExamManualScoringService(req prot.SaveAnswerManualRequest, userID int64) error {
	fileUrl := utils.StripDomain(req.FileUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

	scoring := &models.ExamQuestionUserManualScoring{
		ExamID:     req.ExamId,
		LessonID:   req.LessonId,
		UserID:     userID,
		QuestionID: req.QuestionId,
		Answer:     &req.Answer,
		FileInfo:   fileInfo,
		IsScored:   false,
	}
	return s.repo.SaveExamManualScoringRepository(scoring)
}

// SaveExerciseManualScoringService SaveExerciseManualScoring lưu câu trả lời cần chấm điểm thủ công cho bài tập
func (s *manualScoringService) SaveExerciseManualScoringService(req prot.SaveAnswerManualRequest, userID int64) error {
	fileUrl := utils.StripDomain(req.FileUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

	scoring := &models.ExerciseQuestionUserManualScoring{
		ExerciseID: req.ExerciseId,
		LessonID:   req.LessonId,
		UserID:     userID,
		QuestionID: req.QuestionId,
		Answer:     &req.Answer,
		FileInfo:   fileInfo,
		IsScored:   false,
	}
	return s.repo.SaveExerciseManualScoringRepository(scoring)
}

// SaveHomeworkManualScoringService SaveHomeworkManualScoring lưu câu trả lời cần chấm điểm thủ công cho bài thi
func (s *manualScoringService) SaveHomeworkManualScoringService(req prot.SaveAnswerManualRequest, userID int64) error {
	tx := db.MasterDB.Begin()

	fileUrl := utils.StripDomain(req.FileUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

	scoring := &models.HomeworkQuestionUserManualScoring{
		HomeworkID: req.HomeworkId,
		LessonID:   req.LessonId,
		UserID:     userID,
		QuestionID: req.QuestionId,
		Answer:     &req.Answer,
		FileInfo:   fileInfo,
		IsScored:   false,
	}
	err := s.correctRepo.UpsertHomeworkUserOnCorrect(req.HomeworkId, req.LessonId, userID, req.QuestionId, "manual_scoring")
	if err != nil {
		tx.Rollback()
		return err
	}

	if fileUrl != "" || req.Answer != "" {
		err := s.repo.SaveHomeworkManualScoringRepository(scoring)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()

	// Sau khi lưu câu trả lời manual, cũng ghi log vào homework_user_questions:
	// - Xóa các bản ghi cũ trùng bộ 4 (homework_id, user_id, question_id, lesson_id)
	// - Thêm bản ghi mới với star = 5, ratio_score = -1, weight theo logic cũ

	// Tính weight theo logic cũ từ cloned_questions
	weight, err := s.homeworkUserQuestionService.GetWeightForQuestion(req.HomeworkId, req.QuestionId)
	if err != nil {
		// Nếu lỗi, fallback weight = 1
		weight = 1
	}

	hwqRepo := repositories.NewHomeworkUserQuestionRepository()
	// Xóa các bản ghi cũ cho bộ 4
	if err := hwqRepo.DeleteByKey(req.HomeworkId, userID, req.QuestionId, req.LessonId, nil); err != nil {
		return err
	}

	// Thêm bản ghi mới
	newRecord := &models.HomeworkUserQuestion{
		HomeworkID:     req.HomeworkId,
		LessonID:       req.LessonId,
		UserID:         userID,
		QuestionID:     req.QuestionId,
		RatioScore:     -1,
		IsAllCorrect:   true,
		Star:           5,
		NumberOptions:  0,
		NumberTimeSent: 1,
		Weight:         weight,
	}
	if err := hwqRepo.SaveHomeworkUserQuestion(newRecord, nil); err != nil {
		return err
	}

	return nil
}

func (s *saveScoreManualScoringService) SaveScoreManualScoring(req *prot.SaveScoreManualRequest, scoringBy int64) error {
	for qidStr, score := range req.ScoreList {
		questionID, err := strconv.ParseInt(qidStr, 10, 64)
		if err != nil {
			continue // skip invalid key
		}
		err = s.repo.UpdateManualScoring(req.ExamId, req.UserId, questionID, score, scoringBy)
		if err != nil {
			return err
		}
	}
	// Sau khi chấm điểm xong, cập nhật lại điểm tổng
	score, err := s.examUserService.CalculateExamScoreService(req.ExamId, req.UserId)
	if err != nil {
		return err
	}
	if err := s.examUserService.SaveExamScoreService(req.ExamId, req.UserId, score); err != nil {
		return err
	}

	// Tính toán ratio cho exam
	ratio, err := s.examUserService.CalculateExamRatioService(req.ExamId, req.UserId)
	if err != nil {
		return err
	}
	if err := s.examUserService.SaveExamRatioService(req.ExamId, req.UserId, ratio); err != nil {
		return err
	}

	// Kiểm tra xem còn câu hỏi nào chưa chấm không
	hasUnscoredQuestions, err := s.repo.CheckUnscoredQuestions(req.ExamId, req.UserId)
	if err != nil {
		return err
	}

	// Nếu không còn câu hỏi nào chưa chấm, cập nhật has_manual_scoring = false
	if !hasUnscoredQuestions {
		err = s.repo.UpdateExamUserManualScoringStatus(req.ExamId, req.UserId, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *saveScoreManualScoringService) SaveScoreManualScoringExercise(req *prot.SaveScoreManualRequest, scoringBy int64) error {
	for qidStr, score := range req.ScoreList {
		questionID, err := strconv.ParseInt(qidStr, 10, 64)
		if err != nil {
			continue // skip invalid key
		}
		err = s.repo.UpdateManualScoringExercise(req.ExerciseId, req.UserId, questionID, score, scoringBy)
		if err != nil {
			return err
		}
	}
	// Sau khi chấm điểm xong, cập nhật lại điểm tổng cho exercise
	score, err := s.examUserService.CalculateExerciseScoreService(req.ExerciseId, req.UserId)
	if err != nil {
		return err
	}
	if err := s.examUserService.SaveExerciseScoreService(req.ExerciseId, req.UserId, score); err != nil {
		return err
	}

	// Tính toán ratio cho exercise
	ratio, err := s.examUserService.CalculateExerciseRatioService(req.ExerciseId, req.UserId)
	if err != nil {
		return err
	}
	if err := s.examUserService.SaveExerciseRatioService(req.ExerciseId, req.UserId, ratio); err != nil {
		return err
	}

	// Kiểm tra xem còn câu hỏi nào chưa chấm không
	hasUnscoredQuestions, err := s.repo.CheckUnscoredQuestionsExercise(req.ExerciseId, req.UserId)
	if err != nil {
		return err
	}

	// Nếu không còn câu hỏi nào chưa chấm, cập nhật has_manual_scoring = false
	if !hasUnscoredQuestions {
		err = s.repo.UpdateExerciseUserManualScoringStatus(req.ExerciseId, req.UserId, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveScoreManualScoringHomework chấm điểm thủ công cho homework
func (s *saveScoreManualScoringService) SaveScoreManualScoringHomework(req *prot.SaveScoreManualRequest, scoringBy int64) error {
	for qidStr, score := range req.ScoreList {
		questionID, err := strconv.ParseInt(qidStr, 10, 64)
		if err != nil {
			continue // skip invalid key
		}

		// Cập nhật điểm chấm tay trong bảng homework_question_user_manual_scoring
		if err := s.repo.UpdateManualScoringHomework(req.HomeworkId, req.UserId, questionID, score, scoringBy); err != nil {
			return err
		}

		// Lấy lesson_id tương ứng từ bảng homework_question_user_manual_scoring (nếu có)
		var manual models.HomeworkQuestionUserManualScoring
		var lessonID int64
		if err := db.MasterDB.
			Where("homework_id = ? AND user_id = ? AND question_id = ?", req.HomeworkId, req.UserId, questionID).
			Order("created_at DESC, id DESC").
			First(&manual).Error; err == nil {
			lessonID = manual.LessonID
		}

		// Update ratio_score trong homework_user_questions cho tất cả bản ghi có bộ 4 trùng
		// ratio_score = điểm nhập vào * 100
		homeworkUserQuestionRepo := repositories.NewHomeworkUserQuestionRepository()
		ratioScore := score * 10.0
		if err := homeworkUserQuestionRepo.UpdateRatioScoreByKey(req.HomeworkId, req.UserId, questionID, lessonID, ratioScore, nil); err != nil {
			return err
		}
	}

	// Sau khi chấm điểm xong, cập nhật lại điểm tổng cho homework
	score, err := s.homeworkUserService.CalculateHomeworkScoreService(req.HomeworkId, req.UserId)
	if err != nil {
		return err
	}
	if err := s.homeworkUserService.SaveHomeworkScoreService(req.HomeworkId, req.UserId, score); err != nil {
		return err
	}

	// Tính toán ratio cho homework
	ratio, err := s.homeworkUserService.CalculateHomeworkRatioService(req.HomeworkId, req.UserId)
	if err != nil {
		return err
	}
	if err := s.homeworkUserService.SaveHomeworkRatioService(req.HomeworkId, req.UserId, ratio); err != nil {
		return err
	}

	// Kiểm tra và cập nhật trạng thái has_manual_scoring cho HomeworkUser
	hasUnscored, err := s.repo.CheckUnscoredQuestionsHomework(req.HomeworkId, req.UserId)
	if err != nil {
		return err
	}
	if !hasUnscored {
		if err := s.repo.UpdateHomeworkUserManualScoringStatus(req.HomeworkId, req.UserId, false); err != nil {
			return err
		}
	}

	// Cập nhật status_scoring sau khi chấm điểm thủ công
	if err := s.homeworkUserService.UpdateHomeworkStatusScoringService(req.HomeworkId, req.UserId); err != nil {
		return err
	}

	// Lấy lesson_id từ homework_users để tính toán metrics
	homeworkUserRepo := repositories.NewHomeworkUserRepository()
	homeworkUser, err := homeworkUserRepo.GetByHomeworkAndUser(req.HomeworkId, req.UserId)
	var lessonID int64 = 0
	
	if err != nil {
		// Nếu không tìm thấy homework_users, thử lấy từ manual scoring record đầu tiên
		var manual models.HomeworkQuestionUserManualScoring
		if err := db.MasterDB.
			Where("homework_id = ? AND user_id = ?", req.HomeworkId, req.UserId).
			Order("created_at DESC, id DESC").
			First(&manual).Error; err == nil {
			lessonID = manual.LessonID
			homeworkUser = &models.HomeworkUser{
				LessonID: manual.LessonID,
			}
		}
		// Nếu vẫn không tìm thấy, lessonID = 0 (vẫn tính toán được)
	} else {
		// Lấy lesson_id từ homework_users (có thể = 0)
		lessonID = homeworkUser.LessonID
	}

	// Lấy giá trị cũ từ homework_users trước khi tính toán mới
	var oldExp float64
	var oldStar int64
	if homeworkUser != nil && homeworkUser.ID > 0 {
		oldExp = homeworkUser.Exp
		oldStar = int64(homeworkUser.Star)
	}

	// Tính toán lại exp, star, ratio_score qua service chuyên dụng
	calculationResult, err := s.homeworkCalculationService.CalculateHomeworkMetrics(req.HomeworkId, req.UserId, lessonID)
	if err != nil {
		// Log lỗi nhưng không dừng flow
		return nil
	}

	// Lưu exp và star vào homework_users (không cập nhật updated_at)
	if err := s.homeworkUserService.SaveHomeworkExpService(req.HomeworkId, req.UserId, calculationResult.Exp); err != nil {
		// Log lỗi nhưng không dừng flow
	}
	if err := s.homeworkUserService.SaveHomeworkStarService(req.HomeworkId, req.UserId, calculationResult.Star); err != nil {
		// Log lỗi nhưng không dừng flow
	}

	// Tính change_star và change_exp
	// Nếu chưa có record cũ: change = giá trị mới
	// Nếu đã có record cũ: change = giá trị mới - giá trị cũ (có thể âm nếu mới < cũ)
	var changeStar int64
	var changeExp float64
	if homeworkUser == nil || homeworkUser.ID == 0 {
		// Chưa có record cũ, change = giá trị mới
		changeStar = calculationResult.Star
		changeExp = calculationResult.Exp
	} else {
		// Đã có record cũ, change = giá trị mới - giá trị cũ
		changeStar = calculationResult.Star - oldStar
		changeExp = calculationResult.Exp - oldExp
	}

	// Lưu tổng star và exp vào bảng user_star_exp
	note := fmt.Sprintf("chấm điểm lại homework %d", req.HomeworkId)
	if err := s.userStarExpService.UpdateStarAndExp(req.UserId, changeStar, changeExp, note); err != nil {
		// Log lỗi nhưng không dừng flow
	}

	return nil
}

