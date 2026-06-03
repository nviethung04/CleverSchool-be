package services

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
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
	repo        repositories.ManualScoringRepository
	correctRepo repositories.SaveCorrectHomeworkRepository
}

type saveScoreManualScoringService struct {
	repo              repositories.SaveScoreManualScoringRepository
	examUserService   ExamUserService
	homeworkUserService HomeworkUserService
}

func NewManualScoringService(repo repositories.ManualScoringRepository) ManualScoringService {
	return &manualScoringService{
		repo:        repo,
		correctRepo: repositories.NewSaveCorrectHomeworkRepository(),
	}
}

func NewSaveScoreManualScoringService(repo repositories.SaveScoreManualScoringRepository, examUserService ExamUserService, homeworkUserService HomeworkUserService) SaveScoreManualScoringService {
	return &saveScoreManualScoringService{
		repo:              repo,
		examUserService:   examUserService,
		homeworkUserService: homeworkUserService,
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
        err = s.repo.UpdateManualScoringHomework(req.HomeworkId, req.UserId, questionID, score, scoringBy)
        if err != nil {
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
    return nil
}
