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

type SaveScoreGroupService interface {
	SaveScoreGroupExam(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error)
	SaveScoreGroupHomework(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error)
	SaveScoreGroupExercise(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error)
}

type saveScoreGroupService struct {
	repo                        repositories.SaveScoreGroupRepository
	correctRepo                 repositories.SaveCorrectHomeworkRepository
	clonedQuestionService       ClonedQuestionService
	homeworkUserQuestionService HomeworkUserQuestionService
}

func NewSaveScoreGroupService(repo repositories.SaveScoreGroupRepository, clonedQuestionService ClonedQuestionService) SaveScoreGroupService {
	homeworkUserQuestionRepo := repositories.NewHomeworkUserQuestionRepository()
	return &saveScoreGroupService{
		repo:                        repo,
		correctRepo:                 repositories.NewSaveCorrectHomeworkRepository(),
		clonedQuestionService:       clonedQuestionService,
		homeworkUserQuestionService: NewHomeworkUserQuestionService(homeworkUserQuestionRepo, clonedQuestionService),
	}
}

func (s *saveScoreGroupService) SaveScoreGroupExam(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error) {
	if req.ExamId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	// Lấy danh sách câu hỏi từ cloned_question
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExamId, "exam")
	if err != nil {
		return nil, err
	}
	q, ok := questions[fmt.Sprintf("%v", req.QuestionId)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	// Parse correct_answers.list để lấy đáp án đúng
	var correctAnswers struct {
		List map[string]string `json:"list"`
	}
	if err := json.Unmarshal(q.CorrectAnswers, &correctAnswers); err != nil {
		return nil, fmt.Errorf("Lỗi parse correct_answers.list: %v", err)
	}

	// Parse options.items và options.categories để lấy map id -> content/group_id
	var options struct {
		Items []struct {
			ID              json.Number `json:"id"`
			Text            string      `json:"text"`
			GroupPosition   int64       `json:"group_position"`
			CorrectPosition int64       `json:"correct_position"`
		} `json:"items"`
		Categories []struct {
			ID   json.Number `json:"id"`
			Name string      `json:"name"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.items/categories: %v", err)
	}
	itemIdToText := make(map[int64]string)
	itemIdToGroup := make(map[int64]int64)
	for _, item := range options.Items {
		if id, err := strconv.ParseInt(item.ID.String(), 10, 64); err == nil {
			itemIdToText[id] = item.Text
			itemIdToGroup[id] = item.GroupPosition
		}
	}
	groupIdToName := make(map[int64]string)
	for _, cat := range options.Categories {
		if id, err := strconv.ParseInt(cat.ID.String(), 10, 64); err == nil {
			groupIdToName[id] = cat.Name
		}
	}

	numGroups := len(req.Answers)
	perGroupScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var groupResults []*prot.GroupPair

	for answerIDStr, groupID := range req.Answers {
		var answerID int64
		if answerId, err := strconv.ParseInt(answerIDStr, 10, 64); err == nil {
			answerID = answerId
		}
		correctGroupIDStr := correctAnswers.List[fmt.Sprintf("%v", answerID)]
		correctGroupID, _ := strconv.ParseInt(correctGroupIDStr, 10, 64)
		isCorrect := groupID == correctGroupID
		score := 0.0
		if isCorrect {
			score = perGroupScore
			correctCount++
			totalScore += score
		}
		groupResults = append(groupResults, &prot.GroupPair{
			AnswerId:      answerID,
			GroupId:       groupID,
			AnswerContent: itemIdToText[answerID],
			GroupContent:  groupIdToName[groupID],
			IsCorrect:     isCorrect,
			Score:         score,
		})
	}

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var records []*models.ExamQuestionUserGroup
	for answerIDStr, groupID := range req.Answers {
		var answerID int64
		if answerId, err := strconv.ParseInt(answerIDStr, 10, 64); err == nil {
			answerID = answerId
		}
		correctGroupIDStr := correctAnswers.List[fmt.Sprintf("%v", answerID)]
		correctGroupID, _ := strconv.ParseInt(correctGroupIDStr, 10, 64)
		isCorrect := groupID == correctGroupID
		score := 0.0
		if isCorrect {
			score = perGroupScore
		}
		records = append(records, &models.ExamQuestionUserGroup{
			ExamID:     req.ExamId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			AnswerID:   answerID,
			GroupID:    groupID,
			IsCorrect:  isCorrect,
			Score:      score,
		})
	}
	if err := s.repo.SaveBatchExamQuestionUserGroup(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseGroup{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Groups:       groupResults,
		IsAllCorrect: correctCount == numGroups,
	}, nil
}

func (s *saveScoreGroupService) SaveScoreGroupHomework(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error) {
	if req.HomeworkId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	// Lấy danh sách câu hỏi từ cloned_question
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.HomeworkId, "homework")
	if err != nil {
		return nil, err
	}
	q, ok := questions[fmt.Sprintf("%v", req.QuestionId)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	// Parse correct_answers.list để lấy đáp án đúng
	var correctAnswers struct {
		List map[string]string `json:"list"`
	}
	if err := json.Unmarshal(q.CorrectAnswers, &correctAnswers); err != nil {
		return nil, fmt.Errorf("Lỗi parse correct_answers.list: %v", err)
	}

	// Parse options.items và options.categories để lấy map id -> content/group_id
	var options struct {
		Items []struct {
			ID              json.Number `json:"id"`
			Text            string      `json:"text"`
			GroupPosition   int64       `json:"group_position"`
			CorrectPosition int64       `json:"correct_position"`
		} `json:"items"`
		Categories []struct {
			ID   json.Number `json:"id"`
			Name string      `json:"name"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.items/categories: %v", err)
	}
	itemIdToText := make(map[int64]string)
	itemIdToGroup := make(map[int64]int64)
	for _, item := range options.Items {
		if id, err := strconv.ParseInt(item.ID.String(), 10, 64); err == nil {
			itemIdToText[id] = item.Text
			itemIdToGroup[id] = item.GroupPosition
		}
	}
	groupIdToName := make(map[int64]string)
	for _, cat := range options.Categories {
		if id, err := strconv.ParseInt(cat.ID.String(), 10, 64); err == nil {
			groupIdToName[id] = cat.Name
		}
	}

	numGroups := len(req.Answers)
	perGroupScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var groupResults []*prot.GroupPair

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Kiểm tra và xóa câu trả lời cũ nếu cần
	needSave, err := CheckAndCleanExistingAnswersWithTx(tx, "homework_question_user_groups", req.HomeworkId, userID, req.QuestionId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var records []*models.HomeworkQuestionUserGroup
	for answerIDStr, groupID := range req.Answers {
		var answerID int64
		if answerId, err := strconv.ParseInt(answerIDStr, 10, 64); err == nil {
			answerID = answerId
		}
		correctGroupIDStr := correctAnswers.List[fmt.Sprintf("%v", answerID)]
		correctGroupID, _ := strconv.ParseInt(correctGroupIDStr, 10, 64)
		isCorrect := groupID == correctGroupID
		score := 0.0
		if isCorrect {
			score = perGroupScore
			correctCount++
			totalScore += score
		}
		records = append(records, &models.HomeworkQuestionUserGroup{
			HomeworkID: req.HomeworkId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			AnswerID:   answerID,
			GroupID:    groupID,
			IsCorrect:  isCorrect,
			Score:      score,
		})
		groupResults = append(groupResults, &prot.GroupPair{
			AnswerId:      answerID,
			GroupId:       groupID,
			AnswerContent: itemIdToText[answerID],
			GroupContent:  groupIdToName[groupID],
			IsCorrect:     isCorrect,
			Score:         score,
		})
	}

	isAllCorrect := correctCount == numGroups

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

	// Nếu không cần lưu (đã có câu trả lời đúng hết), trả về kết quả hiện tại + thông tin star/ratio_score/weight/number_time_sent
	if !needSave {
		tx.Commit()
		return &prot.SaveScoreResponseGroup{
			QuestionId:      req.QuestionId,
			TotalScore:      utils.RoundTo2Decimal(totalScore),
			Groups:          groupResults,
			IsAllCorrect:    isAllCorrect,
			Star:            int32(star),
			RatioScore:      ratioScore,
			Weight:          weight,
			NumberTimeSent:  int32(numberTimeSent),
		}, nil
	}

	// Upsert trạng thái hoàn thành nếu đúng hết
	if isAllCorrect {
		err := s.correctRepo.UpsertHomeworkUserOnCorrect(req.HomeworkId, req.LessonId, userID, req.QuestionId, "group")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := s.repo.SaveBatchHomeworkQuestionUserGroup(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseGroup{
		QuestionId:      req.QuestionId,
		TotalScore:      utils.RoundTo2Decimal(totalScore),
		Groups:          groupResults,
		IsAllCorrect:    isAllCorrect,
		Star:            int32(star),
		RatioScore:      ratioScore,
		Weight:          weight,
		NumberTimeSent:  int32(numberTimeSent),
	}, nil
}

func (s *saveScoreGroupService) SaveScoreGroupExercise(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error) {
	if req.ExerciseId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExerciseId, "exercise")
	if err != nil {
		return nil, err
	}
	q, ok := questions[fmt.Sprintf("%v", req.QuestionId)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}
	var correctAnswers struct {
		List map[string]string `json:"list"`
	}
	if err := json.Unmarshal(q.CorrectAnswers, &correctAnswers); err != nil {
		return nil, fmt.Errorf("Lỗi parse correct_answers.list: %v", err)
	}
	var options struct {
		Items []struct {
			ID              json.Number `json:"id"`
			Text            string      `json:"text"`
			GroupPosition   int64       `json:"group_position"`
			CorrectPosition int64       `json:"correct_position"`
		} `json:"items"`
		Categories []struct {
			ID   json.Number `json:"id"`
			Name string      `json:"name"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.items/categories: %v", err)
	}
	itemIdToText := make(map[int64]string)
	groupIdToName := make(map[int64]string)
	for _, item := range options.Items {
		if id, err := strconv.ParseInt(item.ID.String(), 10, 64); err == nil {
			itemIdToText[id] = item.Text
		}
	}
	for _, cat := range options.Categories {
		if id, err := strconv.ParseInt(cat.ID.String(), 10, 64); err == nil {
			groupIdToName[id] = cat.Name
		}
	}
	numGroups := len(req.Answers)
	perGroupScore := 1.0
	totalScore := 0.0
	correctCount := 0
	var groupResults []*prot.GroupPair
	var records []*models.ExerciseQuestionUserGroup
	for answerIDStr, groupID := range req.Answers {
		var answerID int64
		if answerId, err := strconv.ParseInt(answerIDStr, 10, 64); err == nil {
			answerID = answerId
		}
		correctGroupIDStr := correctAnswers.List[fmt.Sprintf("%v", answerID)]
		correctGroupID, _ := strconv.ParseInt(correctGroupIDStr, 10, 64)
		isCorrect := groupID == correctGroupID
		score := 0.0
		if isCorrect {
			score = perGroupScore
			correctCount++
			totalScore += score
		}
		records = append(records, &models.ExerciseQuestionUserGroup{ExerciseID: req.ExerciseId, LessonID: req.LessonId, UserID: userID, QuestionID: req.QuestionId, AnswerID: answerID, GroupID: groupID, IsCorrect: isCorrect, Score: score})
		groupResults = append(groupResults, &prot.GroupPair{AnswerId: answerID, GroupId: groupID, AnswerContent: itemIdToText[answerID], GroupContent: groupIdToName[groupID], IsCorrect: isCorrect, Score: score})
	}
	if err := s.repo.SaveBatchExerciseQuestionUserGroup(records, db.MasterDB); err != nil {
		return nil, err
	}
	return &prot.SaveScoreResponseGroup{QuestionId: req.QuestionId, TotalScore: utils.RoundTo2Decimal(totalScore), Groups: groupResults, IsAllCorrect: correctCount == numGroups}, nil
}
