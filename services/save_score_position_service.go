package services

import (
	"be-lms/database/db"
	_ "be-lms/database/db"
	"be-lms/i18n"
	"be-lms/models"
	_ "be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"encoding/json"
	"fmt"
	_ "math"
	"strconv"
)

type SaveScorePositionService interface {
	SaveScorePositionExam(req *prot.SaveScorePositionRequest, userID int64) (*prot.SaveScoreResponsePosition, error)
	SaveScorePositionHomework(req *prot.SaveScorePositionRequest, userID int64) (*prot.SaveScoreResponsePosition, error)
	SaveScorePositionExercise(req *prot.SaveScorePositionRequest, userID int64) (*prot.SaveScoreResponsePosition, error)
	//SaveScorePositionLevelTest(req *requests.SaveScorePositionRequest, userID int64) (*prot.SaveScoreResponse, error)
}

type saveScorePositionService struct {
	repo                  repositories.SaveScorePositionRepository
	correctRepo           repositories.SaveCorrectHomeworkRepository
	clonedQuestionService ClonedQuestionService
}

func NewSaveScorePositionService(repo repositories.SaveScorePositionRepository, clonedQuestionService ClonedQuestionService) SaveScorePositionService {
	return &saveScorePositionService{
		repo:                  repo,
		correctRepo:           repositories.NewSaveCorrectHomeworkRepository(),
		clonedQuestionService: clonedQuestionService,
	}
}

func (s *saveScorePositionService) SaveScorePositionExam(req *prot.SaveScorePositionRequest, userID int64) (*prot.SaveScoreResponsePosition, error) {
	if req.ExamId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}
	if len(req.Answers) == 0 {
		return &prot.SaveScoreResponsePosition{
			QuestionId:   req.QuestionId,
			TotalScore:   0,
			Answers:      []*prot.AnswerResultPosition{},
			IsAllCorrect: false,
		}, nil
	}
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExamId, "exam")
	if err != nil {
		return nil, err
	}
	q, ok := questions[strconv.FormatInt(req.QuestionId, 10)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}
	// Parse correct_answers.list
	var correctAnswers struct {
		List map[string]string `json:"list"`
	}
	err = json.Unmarshal(q.CorrectAnswers, &correctAnswers)
	if err != nil {
		return nil, fmt.Errorf("Lỗi parse correct_answers: %v", err)
	}

	// Parse options.answers để lấy map position -> group_position đúng
	var options struct {
		Answers []struct {
			GroupPosition   int `json:"group_position"`
			CorrectPosition int `json:"correct_position"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	positionToGroupPos := make(map[int]int)
	for _, ans := range options.Answers {
		positionToGroupPos[ans.CorrectPosition] = ans.GroupPosition
	}

	numAnswers := len(req.Answers)
	perAnswerScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var answerResults []*prot.AnswerResultPosition
	var records []*models.ExamQuestionUserPosition
	for positionStr, groupPos := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}
		correctGroupPos, ok := positionToGroupPos[int(position)]
		isCorrect := false
		if ok && correctGroupPos == int(groupPos) {
			isCorrect = true
			correctCount++
			totalScore += perAnswerScore
		}
		answerResults = append(answerResults, &prot.AnswerResultPosition{
			SortPosition:            int32(position),
			AnswerGroupPosition:     int64(groupPos),
			TrueAnswerGroupPosition: int64(correctGroupPos),
			TrueAnswerContent:       "",
			IsCorrect:               isCorrect,
			Score:                   perAnswerScore,
		})
		records = append(records, &models.ExamQuestionUserPosition{
			ExamID:              req.ExamId,
			LessonID:            req.LessonId,
			UserID:              userID,
			QuestionID:          req.QuestionId,
			AnswerGroupPosition: int64(groupPos),
			SortPosition:        int(position),
			IsCorrect:           isCorrect,
			Score:               perAnswerScore,
		})
	}
	// Lưu DB
	err = s.repo.SaveBatchExamQuestionUserPositions(records, db.MasterDB)
	if err != nil {
		return nil, fmt.Errorf("Lỗi lưu kết quả position exam: %v", err)
	}
	return &prot.SaveScoreResponsePosition{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Answers:      answerResults,
		IsAllCorrect: correctCount == numAnswers,
	}, nil
}

func (s *saveScorePositionService) SaveScorePositionHomework(req *prot.SaveScorePositionRequest, userID int64) (*prot.SaveScoreResponsePosition, error) {
	if req.HomeworkId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}
	if len(req.Answers) == 0 {
		return &prot.SaveScoreResponsePosition{
			QuestionId:   req.QuestionId,
			TotalScore:   0,
			Answers:      []*prot.AnswerResultPosition{},
			IsAllCorrect: false,
		}, nil
	}
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.HomeworkId, "homework")
	if err != nil {
		return nil, err
	}
	q, ok := questions[strconv.FormatInt(req.QuestionId, 10)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}
	// Parse options.answers để lấy map position -> group_position đúng
	var options struct {
		Answers []struct {
			GroupPosition   int `json:"group_position"`
			CorrectPosition int `json:"correct_position"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	positionToGroupPos := make(map[int]int)
	for _, ans := range options.Answers {
		positionToGroupPos[ans.CorrectPosition] = ans.GroupPosition
	}

	numAnswers := len(req.Answers)
	perAnswerScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var answerResults []*prot.AnswerResultPosition
	var records []*models.HomeworkQuestionUserPosition
	for positionStr, groupPos := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}
		correctGroupPos, ok := positionToGroupPos[int(position)]
		isCorrect := false
		if ok && correctGroupPos == int(groupPos) {
			isCorrect = true
			correctCount++
			totalScore += perAnswerScore
		}
		answerResults = append(answerResults, &prot.AnswerResultPosition{
			SortPosition:            int32(position),
			AnswerGroupPosition:     int64(groupPos),
			TrueAnswerGroupPosition: int64(correctGroupPos),
			TrueAnswerContent:       "",
			IsCorrect:               isCorrect,
			Score:                   perAnswerScore,
		})
		records = append(records, &models.HomeworkQuestionUserPosition{
			HomeworkID:          req.HomeworkId,
			LessonID:            req.LessonId,
			UserID:              userID,
			QuestionID:          req.QuestionId,
			AnswerGroupPosition: int64(groupPos),
			SortPosition:        int(position),
			IsCorrect:           isCorrect,
			Score:               perAnswerScore,
		})
	}

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Kiểm tra và xóa câu trả lời cũ nếu cần
	needSave, err := CheckAndCleanExistingAnswersWithTx(tx, "homework_question_user_positions", req.HomeworkId, userID, req.QuestionId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Nếu không cần lưu (đã có câu trả lời đúng hết), trả về kết quả hiện tại
	if !needSave {
		tx.Commit()
		return &prot.SaveScoreResponsePosition{
			QuestionId:   req.QuestionId,
			TotalScore:   utils.RoundTo2Decimal(totalScore),
			Answers:      answerResults,
			IsAllCorrect: correctCount == numAnswers,
		}, nil
	}

	// Upsert trạng thái hoàn thành nếu đúng hết
	if correctCount == numAnswers {
		err := s.correctRepo.UpsertHomeworkUserOnCorrect(req.HomeworkId, req.LessonId, userID, req.QuestionId, "position")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// Lưu DB
	err = s.repo.SaveBatchHomeworkQuestionUserPositions(records, tx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Lỗi lưu kết quả position homework: %v", err)
	}
	tx.Commit()
	return &prot.SaveScoreResponsePosition{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Answers:      answerResults,
		IsAllCorrect: correctCount == numAnswers,
	}, nil
}

func (s *saveScorePositionService) SaveScorePositionExercise(req *prot.SaveScorePositionRequest, userID int64) (*prot.SaveScoreResponsePosition, error) {
	if req.ExerciseId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}
	if len(req.Answers) == 0 {
		return &prot.SaveScoreResponsePosition{QuestionId: req.QuestionId, TotalScore: 0, Answers: []*prot.AnswerResultPosition{}, IsAllCorrect: false}, nil
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
			GroupPosition   int `json:"group_position"`
			CorrectPosition int `json:"correct_position"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	positionToGroupPos := make(map[int]int)
	for _, ans := range options.Answers {
		positionToGroupPos[ans.CorrectPosition] = ans.GroupPosition
	}
	numAnswers := len(req.Answers)
	perAnswerScore := 1.0
	totalScore := 0.0
	correctCount := 0
	var answerResults []*prot.AnswerResultPosition
	var records []*models.ExerciseQuestionUserPosition
	for positionStr, groupPos := range req.Answers {
		var position int
		if pos, err := strconv.Atoi(positionStr); err == nil {
			position = pos
		}
		cg, ok := positionToGroupPos[int(position)]
		isCorrect := false
		if ok && cg == int(groupPos) {
			isCorrect = true
			correctCount++
			totalScore += perAnswerScore
		}
		answerResults = append(answerResults, &prot.AnswerResultPosition{SortPosition: int32(position), AnswerGroupPosition: int64(groupPos), TrueAnswerGroupPosition: int64(cg), TrueAnswerContent: "", IsCorrect: isCorrect, Score: perAnswerScore})
		records = append(records, &models.ExerciseQuestionUserPosition{ExerciseID: req.ExerciseId, LessonID: req.LessonId, UserID: userID, QuestionID: req.QuestionId, AnswerGroupPosition: int64(groupPos), SortPosition: int(position), IsCorrect: isCorrect, Score: perAnswerScore})
	}
	if err := s.repo.SaveBatchExerciseQuestionUserPositions(records, db.MasterDB); err != nil {
		return nil, fmt.Errorf("Lỗi lưu kết quả position exercise: %v", err)
	}
	return &prot.SaveScoreResponsePosition{QuestionId: req.QuestionId, TotalScore: utils.RoundTo2Decimal(totalScore), Answers: answerResults, IsAllCorrect: correctCount == numAnswers}, nil
}
