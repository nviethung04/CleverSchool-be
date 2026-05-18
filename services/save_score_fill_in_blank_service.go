package services

import (
	"be-Clever School/database/db"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type SaveScoreFillInBlankService interface {
	SaveScoreFillInBlankExam(req *prot.SaveScoreFillInBlankRequest, userID int64) (*prot.SaveScoreResponseFillInBlank, error)
	SaveScoreFillInBlankHomework(req *prot.SaveScoreFillInBlankRequest, userID int64) (*prot.SaveScoreResponseFillInBlank, error)
	SaveScoreFillInBlankExercise(req *prot.SaveScoreFillInBlankRequest, userID int64) (*prot.SaveScoreResponseFillInBlank, error)
	//SaveScoreFillInBlankLevelTest(req *requests.SaveScoreFillInBlankRequest, userID int64) (*prot.SaveScoreResponse, error)
}

type saveScoreFillInBlankService struct {
	repo                        repositories.SaveScoreFillInBlankRepository
	correctRepo                 repositories.SaveCorrectHomeworkRepository
	clonedQuestionService       ClonedQuestionService
	homeworkUserQuestionService HomeworkUserQuestionService
}

func NewSaveScoreFillInBlankService(repo repositories.SaveScoreFillInBlankRepository, clonedQuestionService ClonedQuestionService) SaveScoreFillInBlankService {
	// Tạo repository cho homework user question service (dùng chung cho nhiều dạng câu hỏi)
	homeworkUserQuestionRepo := repositories.NewHomeworkUserQuestionRepository()
	return &saveScoreFillInBlankService{
		repo:                        repo,
		correctRepo:                 repositories.NewSaveCorrectHomeworkRepository(),
		clonedQuestionService:       clonedQuestionService,
		homeworkUserQuestionService: NewHomeworkUserQuestionService(homeworkUserQuestionRepo, clonedQuestionService),
	}
}


func (s *saveScoreFillInBlankService) SaveScoreFillInBlankExam(req *prot.SaveScoreFillInBlankRequest, userID int64) (*prot.SaveScoreResponseFillInBlank, error) {
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

	// Parse options.answers để lấy map position -> đáp án đúng
	var options struct {
		Answers []struct {
			Content         string `json:"text"`
			CorrectPosition int    `json:"correct_position"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	positionToContent := make(map[int]string)
	for _, ans := range options.Answers {
		positionToContent[ans.CorrectPosition] = strings.TrimSpace(strings.ToLower(ans.Content))
	}

	numAnswers := len(req.Answers)
	perAnswerScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var answerResults []*prot.AnswerResultFillInBlank

	for positionStr, content := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}
		cleanUser := strings.TrimSpace(strings.ToLower(content))
		trueContent := positionToContent[position]
		isCorrect := cleanUser == trueContent

		score := 0.0
		if isCorrect {
			score = perAnswerScore
			correctCount++
			totalScore += score
		}

		answerResults = append(answerResults, &prot.AnswerResultFillInBlank{
			SortPosition: int32(position),
			Answer:       content,
			TrueAnswer:   trueContent,
			IsCorrect:    isCorrect,
			Score:        score,
		})
	}

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var records []*models.ExamQuestionUserFillInBlank
	for positionStr, content := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}
		trueContent := positionToContent[position]
		cleanUser := strings.TrimSpace(strings.ToLower(content))
		isCorrect := cleanUser == trueContent
		score := 0.0
		if isCorrect {
			score = perAnswerScore
		}
		records = append(records, &models.ExamQuestionUserFillInBlank{
			ExamID:       req.ExamId,
			LessonID:     req.LessonId,
			UserID:       userID,
			QuestionID:   req.QuestionId,
			SortPosition: position,
			Answer:       content,
			IsCorrect:    isCorrect,
			Score:        score,
		})
	}
	if err := s.repo.SaveBatchExamQuestionUserFillInBlanks(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseFillInBlank{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Answers:      answerResults,
		IsAllCorrect: correctCount == numAnswers,
	}, nil
}

func (s *saveScoreFillInBlankService) SaveScoreFillInBlankHomework(req *prot.SaveScoreFillInBlankRequest, userID int64) (*prot.SaveScoreResponseFillInBlank, error) {
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

	// Parse options.answers để lấy map position -> đáp án đúng
	var options struct {
		Answers []struct {
			Content         string `json:"text"`
			CorrectPosition int    `json:"correct_position"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	positionToContent := make(map[int]string)
	for _, ans := range options.Answers {
		positionToContent[ans.CorrectPosition] = strings.TrimSpace(strings.ToLower(ans.Content))
	}

	numAnswers := len(req.Answers)
	perAnswerScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var answerResults []*prot.AnswerResultFillInBlank

	for positionStr, content := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}
		cleanUser := strings.TrimSpace(strings.ToLower(content))
		trueContent := positionToContent[position]
		isCorrect := cleanUser == trueContent

		score := 0.0
		if isCorrect {
			score = perAnswerScore
			correctCount++
			totalScore += score
		}

		answerResults = append(answerResults, &prot.AnswerResultFillInBlank{
			SortPosition: int32(position),
			Answer:       content,
			TrueAnswer:   trueContent,
			IsCorrect:    isCorrect,
			Score:        score,
		})
	}

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Kiểm tra và xóa câu trả lời cũ nếu cần
	needSave, err := CheckAndCleanExistingAnswersWithTx(tx, "homework_question_user_fill_in_blanks", req.HomeworkId, userID, req.QuestionId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Nếu không cần lưu (đã có câu trả lời đúng hết), query từ DB để lấy star, ratio_score, weight, number_time_sent
	if !needSave {
		var star, numberTimeSent int
		var ratioScore, weight float64
		
		// Query bản ghi có is_all_correct = true để lấy star, ratio_score, weight, number_time_sent
		if req.HomeworkId != 0 {
			existingRecord, err := s.homeworkUserQuestionService.GetHomeworkUserQuestionByCorrect(req.HomeworkId, userID, req.QuestionId, req.LessonId)
			if err == nil && existingRecord != nil {
				star = existingRecord.Star
				ratioScore = existingRecord.RatioScore
				weight = existingRecord.Weight
				numberTimeSent = existingRecord.NumberTimeSent
			}
		}
		
		tx.Commit()
		return &prot.SaveScoreResponseFillInBlank{
			QuestionId:      req.QuestionId,
			TotalScore:      utils.RoundTo2Decimal(totalScore),
			Answers:         answerResults,
			IsAllCorrect:    correctCount == numAnswers,
			Star:            int32(star),
			RatioScore:      ratioScore,
			Weight:          weight,
			NumberTimeSent:  int32(numberTimeSent),
		}, nil
	}

	var records []*models.HomeworkQuestionUserFillInBlank
	for positionStr, content := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}

		trueContent := positionToContent[position]
		cleanUser := strings.TrimSpace(strings.ToLower(content))
		isCorrect := cleanUser == trueContent
		score := 0.0
		if isCorrect {
			score = perAnswerScore
		}
		records = append(records, &models.HomeworkQuestionUserFillInBlank{
			HomeworkID:   req.HomeworkId,
			LessonID:     req.LessonId,
			UserID:       userID,
			QuestionID:   req.QuestionId,
			SortPosition: position,
			Answer:       content,
			IsCorrect:    isCorrect,
			Score:        score,
		})
	}

	isAllCorrect := correctCount == numAnswers

	// Lưu vào homework_user_questions và lấy star, ratioScore, weight, numberTimeSent
	var star, numberTimeSent int
	var ratioScore, weight float64
	if req.HomeworkId != 0 {
		var err error
		star, ratioScore, weight, numberTimeSent, err = s.homeworkUserQuestionService.SaveHomeworkUserQuestion(req.HomeworkId, userID, req.QuestionId, req.LessonId, isAllCorrect, tx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Upsert trạng thái hoàn thành nếu đúng hết
	if isAllCorrect {
		err := s.correctRepo.UpsertHomeworkUserOnCorrect(req.HomeworkId, req.LessonId, userID, req.QuestionId, "fill_in_blank")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := s.repo.SaveBatchHomeworkQuestionUserFillInBlanks(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseFillInBlank{
		QuestionId:      req.QuestionId,
		TotalScore:      utils.RoundTo2Decimal(totalScore),
		Answers:         answerResults,
		IsAllCorrect:    isAllCorrect,
		Star:            int32(star),
		RatioScore:      ratioScore,
		Weight:          weight,
		NumberTimeSent:  int32(numberTimeSent),
	}, nil
}

func (s *saveScoreFillInBlankService) SaveScoreFillInBlankExercise(req *prot.SaveScoreFillInBlankRequest, userID int64) (*prot.SaveScoreResponseFillInBlank, error) {
	if req.ExerciseId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExerciseId, "exercise")
	if err != nil {
		return nil, err
	}
	q, ok := questions[strconv.FormatInt(req.QuestionId, 10)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}
	var options struct {
		Answers []struct {
			Content         string `json:"text"`
			CorrectPosition int    `json:"correct_position"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	positionToContent := make(map[int]string)
	for _, ans := range options.Answers {
		positionToContent[ans.CorrectPosition] = strings.TrimSpace(strings.ToLower(ans.Content))
	}

	numAnswers := len(req.Answers)
	perAnswerScore := 1.0
	totalScore := 0.0
	correctCount := 0
	var answerResults []*prot.AnswerResultFillInBlank
	for positionStr, content := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}
		cleanUser := strings.TrimSpace(strings.ToLower(content))
		trueContent := positionToContent[position]
		isCorrect := cleanUser == trueContent
		score := 0.0
		if isCorrect {
			score = perAnswerScore
			correctCount++
			totalScore += score
		}
		answerResults = append(answerResults, &prot.AnswerResultFillInBlank{SortPosition: int32(position), Answer: content, TrueAnswer: trueContent, IsCorrect: isCorrect, Score: score})
	}
	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var records []*models.ExerciseQuestionUserFillInBlank
	for positionStr, content := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}
		trueContent := positionToContent[position]
		cleanUser := strings.TrimSpace(strings.ToLower(content))
		isCorrect := cleanUser == trueContent
		score := 0.0
		if isCorrect {
			score = perAnswerScore
		}
		records = append(records, &models.ExerciseQuestionUserFillInBlank{ExerciseID: req.ExerciseId, LessonID: req.LessonId, UserID: userID, QuestionID: req.QuestionId, SortPosition: position, Answer: content, IsCorrect: isCorrect, Score: score})
	}
	if err := s.repo.SaveBatchExerciseQuestionUserFillInBlanks(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return &prot.SaveScoreResponseFillInBlank{QuestionId: req.QuestionId, TotalScore: utils.RoundTo2Decimal(totalScore), Answers: answerResults, IsAllCorrect: correctCount == numAnswers}, nil
}
