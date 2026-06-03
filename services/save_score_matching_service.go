package services

import (
	"be-lms/database/db"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"encoding/json"
	"fmt"
	"strconv"
)

type SaveScoreMatchingService interface {
	SaveScoreMatchingExam(req *prot.SaveScoreMatchingRequest, userID int64) (*prot.SaveScoreResponseMatching, error)
	SaveScoreMatchingHomework(req *prot.SaveScoreMatchingRequest, userID int64) (*prot.SaveScoreResponseMatching, error)
	SaveScoreMatchingLevelTest(req *prot.SaveScoreMatchingRequest, userID int64) (*prot.SaveScoreResponseMatching, error)
	SaveScoreMatchingExercise(req *prot.SaveScoreMatchingRequest, userID int64) (*prot.SaveScoreResponseMatching, error)
}

type saveScoreMatchingService struct {
	repo                  repositories.SaveScoreMatchingRepository
	correctRepo           repositories.SaveCorrectHomeworkRepository
	clonedQuestionService ClonedQuestionService
}

func NewSaveScoreMatchingService(repo repositories.SaveScoreMatchingRepository, clonedQuestionService ClonedQuestionService) SaveScoreMatchingService {
	return &saveScoreMatchingService{
		repo:                  repo,
		correctRepo:           repositories.NewSaveCorrectHomeworkRepository(),
		clonedQuestionService: clonedQuestionService,
	}
}

func (s *saveScoreMatchingService) SaveScoreMatchingExam(req *prot.SaveScoreMatchingRequest, userID int64) (*prot.SaveScoreResponseMatching, error) {
	if req.ExamId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	// Lấy danh sách câu hỏi từ cloned_question
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExamId, "exam")
	if err != nil {
		return nil, err
	}
	q, ok := questions[strconv.FormatInt(req.QuestionId, 10)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	// Parse options.sources và options.targets để lấy map id -> content
	var options struct {
		Sources []struct {
			ID      json.Number `json:"id"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"sources"`
		Targets []struct {
			ID      json.Number `json:"id"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.sources/targets: %v", err)
	}
	idToSourceContent := make(map[int64]string)
	idToTargetContent := make(map[int64]string)
	for _, src := range options.Sources {
		if id, err := strconv.ParseInt(src.ID.String(), 10, 64); err == nil {
			idToSourceContent[id] = src.Content.Text
		}
	}
	for _, tgt := range options.Targets {
		if id, err := strconv.ParseInt(tgt.ID.String(), 10, 64); err == nil {
			idToTargetContent[id] = tgt.Content.Text
		}
	}

	numMatches := len(req.Answers)
	perMatchScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var matchResults []*prot.MatchingPair

	for firstItemIDStr, secondItemID := range req.Answers {
		var firstItemID int64

		if firstId, err := strconv.ParseInt(firstItemIDStr, 10, 64); err == nil {
			firstItemID = firstId
		}

		isCorrect := firstItemID == secondItemID
		score := 0.0
		if isCorrect {
			score = perMatchScore
			correctCount++
			totalScore += score
		}
		matchResults = append(matchResults, &prot.MatchingPair{
			FirstItemId:       firstItemID,
			SecondItemId:      secondItemID,
			FirstItemContent:  idToSourceContent[firstItemID],
			SecondItemContent: idToTargetContent[secondItemID],
			IsCorrect:         isCorrect,
			Score:             score,
		})
	}

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var records []*models.ExamQuestionUserMatching
	for firstItemIDStr, secondItemID := range req.Answers {
		var firstItemID int64

		if firstId, err := strconv.ParseInt(firstItemIDStr, 10, 64); err == nil {
			firstItemID = firstId
		}

		isCorrect := firstItemID == secondItemID
		score := 0.0
		if isCorrect {
			score = perMatchScore
			totalScore += score
			correctCount++
		}
		records = append(records, &models.ExamQuestionUserMatching{
			ExamID:       req.ExamId,
			LessonID:     req.LessonId,
			UserID:       userID,
			QuestionID:   req.QuestionId,
			FirstItemID:  firstItemID,
			SecondItemID: secondItemID,
			IsCorrect:    isCorrect,
			Score:        score,
		})
		matchResults = append(matchResults, &prot.MatchingPair{
			FirstItemId:       firstItemID,
			SecondItemId:      secondItemID,
			FirstItemContent:  idToSourceContent[firstItemID],
			SecondItemContent: idToTargetContent[secondItemID],
			IsCorrect:         isCorrect,
			Score:             score,
		})
	}
	if err := s.repo.SaveBatchExamQuestionUserMatching(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseMatching{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Matches:      matchResults,
		IsAllCorrect: correctCount == numMatches,
	}, nil
}

func (s *saveScoreMatchingService) SaveScoreMatchingHomework(req *prot.SaveScoreMatchingRequest, userID int64) (*prot.SaveScoreResponseMatching, error) {
	if req.HomeworkId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	// Lấy danh sách câu hỏi từ cloned_question
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.HomeworkId, "homework")
	if err != nil {
		return nil, err
	}
	q, ok := questions[strconv.FormatInt(req.QuestionId, 10)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	// Parse options.sources và options.targets để lấy map id -> content
	var options struct {
		Sources []struct {
			ID      json.Number `json:"id"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"sources"`
		Targets []struct {
			ID      json.Number `json:"id"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.sources/targets: %v", err)
	}
	idToSourceContent := make(map[int64]string)
	idToTargetContent := make(map[int64]string)
	for _, src := range options.Sources {
		if id, err := strconv.ParseInt(src.ID.String(), 10, 64); err == nil {
			idToSourceContent[id] = src.Content.Text
		}
	}
	for _, tgt := range options.Targets {
		if id, err := strconv.ParseInt(tgt.ID.String(), 10, 64); err == nil {
			idToTargetContent[id] = tgt.Content.Text
		}
	}

	numMatches := len(req.Answers)
	perMatchScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var matchResults []*prot.MatchingPair

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Kiểm tra và xóa câu trả lời cũ nếu cần
	needSave, err := CheckAndCleanExistingAnswersWithTx(tx, "homework_question_user_matchings", req.HomeworkId, userID, req.QuestionId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var records []*models.HomeworkQuestionUserMatching
	for firstItemIDStr, secondItemID := range req.Answers {
		var firstItemID int64

		if firstId, err := strconv.ParseInt(firstItemIDStr, 10, 64); err == nil {
			firstItemID = firstId
		}

		isCorrect := firstItemID == secondItemID
		score := 0.0
		if isCorrect {
			score = perMatchScore
			totalScore += score
			correctCount++
		}
		records = append(records, &models.HomeworkQuestionUserMatching{
			HomeworkID:   req.HomeworkId,
			LessonID:     req.LessonId,
			UserID:       userID,
			QuestionID:   req.QuestionId,
			FirstItemID:  firstItemID,
			SecondItemID: secondItemID,
			IsCorrect:    isCorrect,
			Score:        score,
		})
		matchResults = append(matchResults, &prot.MatchingPair{
			FirstItemId:       firstItemID,
			SecondItemId:      secondItemID,
			FirstItemContent:  idToSourceContent[firstItemID],
			SecondItemContent: idToTargetContent[secondItemID],
			IsCorrect:         isCorrect,
			Score:             score,
		})
	}

	// Nếu không cần lưu (đã có câu trả lời đúng hết), trả về kết quả hiện tại
	if !needSave {
		tx.Commit()
		return &prot.SaveScoreResponseMatching{
			QuestionId:   req.QuestionId,
			TotalScore:   utils.RoundTo2Decimal(totalScore),
			Matches:      matchResults,
			IsAllCorrect: correctCount == numMatches,
		}, nil
	}

	// Upsert trạng thái hoàn thành nếu đúng hết
	if correctCount == numMatches {
		err := s.correctRepo.UpsertHomeworkUserOnCorrect(req.HomeworkId, req.LessonId, userID, req.QuestionId, "matching")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// Lưu kết quả vào DB
	if err := s.repo.SaveBatchHomeworkQuestionUserMatching(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseMatching{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Matches:      matchResults,
		IsAllCorrect: correctCount == numMatches,
	}, nil
}

func (s *saveScoreMatchingService) SaveScoreMatchingLevelTest(req *prot.SaveScoreMatchingRequest, userID int64) (*prot.SaveScoreResponseMatching, error) {
	if req.LevelTestId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	// Logic tương tự như SaveScoreMatchingExam
	// ... (implement tương tự)
	return nil, nil
}

func (s *saveScoreMatchingService) SaveScoreMatchingExercise(req *prot.SaveScoreMatchingRequest, userID int64) (*prot.SaveScoreResponseMatching, error) {
	if req.ExerciseId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	// Lấy danh sách câu hỏi từ cloned_question
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExerciseId, "exercise")
	if err != nil {
		return nil, err
	}
	q, ok := questions[strconv.FormatInt(req.QuestionId, 10)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	// Parse options.sources và options.targets để lấy map id -> content
	var options struct {
		Sources []struct {
			ID      json.Number `json:"id"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"sources"`
		Targets []struct {
			ID      json.Number `json:"id"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.sources/targets: %v", err)
	}
	idToSourceContent := make(map[int64]string)
	idToTargetContent := make(map[int64]string)
	for _, src := range options.Sources {
		if id, err := strconv.ParseInt(src.ID.String(), 10, 64); err == nil {
			idToSourceContent[id] = src.Content.Text
		}
	}
	for _, tgt := range options.Targets {
		if id, err := strconv.ParseInt(tgt.ID.String(), 10, 64); err == nil {
			idToTargetContent[id] = tgt.Content.Text
		}
	}

	numMatches := len(req.Answers)
	perMatchScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var matchResults []*prot.MatchingPair

	for firstItemIDStr, secondItemID := range req.Answers {
		var firstItemID int64

		if firstId, err := strconv.ParseInt(firstItemIDStr, 10, 64); err == nil {
			firstItemID = firstId
		}
		isCorrect := firstItemID == secondItemID
		score := 0.0
		if isCorrect {
			score = perMatchScore
			correctCount++
			totalScore += score
		}
		matchResults = append(matchResults, &prot.MatchingPair{
			FirstItemId:       firstItemID,
			SecondItemId:      secondItemID,
			FirstItemContent:  idToSourceContent[firstItemID],
			SecondItemContent: idToTargetContent[secondItemID],
			IsCorrect:         isCorrect,
			Score:             score,
		})
	}

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var records []*models.ExerciseQuestionUserMatching
	for firstItemIDStr, secondItemID := range req.Answers {
		var firstItemID int64

		if firstId, err := strconv.ParseInt(firstItemIDStr, 10, 64); err == nil {
			firstItemID = firstId
		}
		isCorrect := firstItemID == secondItemID
		score := 0.0
		if isCorrect {
			score = perMatchScore
			totalScore += score
			correctCount++
		}
		records = append(records, &models.ExerciseQuestionUserMatching{
			ExerciseID:   req.ExerciseId,
			LessonID:     req.LessonId,
			UserID:       userID,
			QuestionID:   req.QuestionId,
			FirstItemID:  firstItemID,
			SecondItemID: secondItemID,
			IsCorrect:    isCorrect,
			Score:        score,
		})
		matchResults = append(matchResults, &prot.MatchingPair{
			FirstItemId:       firstItemID,
			SecondItemId:      secondItemID,
			FirstItemContent:  idToSourceContent[firstItemID],
			SecondItemContent: idToTargetContent[secondItemID],
			IsCorrect:         isCorrect,
			Score:             score,
		})
	}
	if err := s.repo.SaveBatchExerciseQuestionUserMatching(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseMatching{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Matches:      matchResults,
		IsAllCorrect: correctCount == numMatches,
	}, nil
}
