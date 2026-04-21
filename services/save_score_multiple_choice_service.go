package services

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type SaveScoreMultipleChoiceService interface {
	SaveScoreMultipleChoiceExam(req *prot.SaveScoreMultipleChoiceRequest, userID int64) (*prot.SaveScoreResponse, error)
	SaveScoreMultipleChoiceHomework(req *prot.SaveScoreMultipleChoiceRequest, userID int64) (*prot.SaveScoreResponse, error)
	SaveScoreMultipleChoiceLevelTest(req *prot.SaveScoreMultipleChoiceRequest, userID int64) (*prot.SaveScoreResponse, error)
	SaveScoreMultipleChoiceExercise(req *prot.SaveScoreMultipleChoiceRequest, userID int64) (*prot.SaveScoreResponse, error)
}

type saveScoreMultipleChoiceService struct {
	repo                        repositories.SaveScoreMultipleChoiceRepository
	correctRepo                 repositories.SaveCorrectHomeworkRepository
	clonedQuestionService       ClonedQuestionService
	homeworkUserQuestionService HomeworkUserQuestionService
}

func NewSaveScoreMultipleChoiceService(repo repositories.SaveScoreMultipleChoiceRepository, clonedQuestionService ClonedQuestionService) SaveScoreMultipleChoiceService {
	homeworkUserQuestionRepo := repositories.NewHomeworkUserQuestionRepository()
	return &saveScoreMultipleChoiceService{
		repo:                        repo,
		correctRepo:                 repositories.NewSaveCorrectHomeworkRepository(),
		clonedQuestionService:       clonedQuestionService,
		homeworkUserQuestionService: NewHomeworkUserQuestionService(homeworkUserQuestionRepo, clonedQuestionService),
	}
}

// Dùng chung để lấy trueAnswer, userAnswer, tính điểm và đúng/sai
func (s *saveScoreMultipleChoiceService) prepareScoreData(req *prot.SaveScoreMultipleChoiceRequest) ([]*dto.AnswerWithScoreMultipleChoice, []*models.Answer, bool, float64, error) {
	if req.QuestionId == 0 {
		return nil, nil, false, 0, errors.New("missing question_id or answer_ids")
	}

	var trueAnswers []*dto.AnswerWithScoreMultipleChoice
	var err error

	switch {
	case req.ExamId != 0:
		trueAnswers, err = s.repo.GetCorrectAnswersFromExamMultipleChoice(req.ExamId, req.QuestionId)
	case req.LevelTestId != 0:
		trueAnswers, err = s.repo.GetCorrectAnswersFromLevelTestMultipleChoice(req.LevelTestId, req.QuestionId)
	case req.HomeworkId != 0:
		trueAnswers, err = s.repo.GetCorrectAnswersFromHomeworkMultipleChoice(req.HomeworkId, req.QuestionId)
	default:
		err = errors.New("must provide exam_id, level_test_id, or homework_id")
	}
	if err != nil {
		return nil, nil, false, 0, err
	}

	userAnswers, err := s.repo.GetAnswersByIDsMultipleChoice(req.AnswerIds, req.QuestionId)
	if err != nil {
		return nil, nil, false, 0, err
	}

	// Kiểm tra số lượng đáp án có khớp không
	if len(userAnswers) != len(trueAnswers) {
		return trueAnswers, userAnswers, false, 0, nil
	}

	// Tạo map để so sánh nhanh
	trueAnswerMap := make(map[int64]bool)
	for _, ans := range trueAnswers {
		trueAnswerMap[ans.ID] = true
	}

	// Kiểm tra từng đáp án của user có trong đáp án đúng không
	isCorrect := true
	for _, ans := range userAnswers {
		if !trueAnswerMap[ans.ID] {
			isCorrect = false
			break
		}
	}

	score := 0.0
	if isCorrect && len(trueAnswers) > 0 {
		score = trueAnswers[0].Score // Lấy điểm từ đáp án đúng đầu tiên
	}

	return trueAnswers, userAnswers, isCorrect, score, nil
}

func (s *saveScoreMultipleChoiceService) SaveScoreMultipleChoiceExam(req *prot.SaveScoreMultipleChoiceRequest, userID int64) (*prot.SaveScoreResponse, error) {
	if req.ExamId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExamId, "exam")
	if err != nil {
		return nil, err
	}
	q, ok := questions[strconv.FormatInt(req.QuestionId, 10)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	var raw struct {
		Ids  []interface{}     `json:"ids"`
		List map[string]string `json:"list"`
	}

	err = json.Unmarshal(q.CorrectAnswers, &raw)
	if err != nil {
		return nil, fmt.Errorf("Dữ liệu correct_answers không đúng định dạng: %v", err)
	}
	correctAnswers := struct {
		Ids  []string
		List map[string]string
	}{
		List: raw.List,
	}

	for _, v := range raw.Ids {
		switch val := v.(type) {
		case float64:
			correctAnswers.Ids = append(correctAnswers.Ids, fmt.Sprintf("%.0f", val))
		case string:
			correctAnswers.Ids = append(correctAnswers.Ids, val)
		default:
			correctAnswers.Ids = append(correctAnswers.Ids, fmt.Sprint(val))
		}
	}

	var correctList []string
	if len(correctAnswers.Ids) > 0 {
		correctList = correctAnswers.Ids
	} else if len(correctAnswers.List) > 0 {
		for k := range correctAnswers.List {
			correctList = append(correctList, k)
		}
	} else {
		return nil, fmt.Errorf("Không tìm thấy đáp án đúng trong correct_answers")
	}
	// Parse options.answers để lấy map answer_id -> content
	var options struct {
		Answers []struct {
			ID      json.Number `json:"id"`
			Content string      `json:"text"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	idToContent := make(map[string]string)
	for _, ans := range options.Answers {
		idToContent[ans.ID.String()] = ans.Content
	}
	perAnswerScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var records []*models.ExamQuestionUser
	isAllCorrect := true
	score := 0.0
	for _, id := range req.AnswerIds {
		isCorrect := false
		idStr := strconv.FormatInt(id, 10)
		for _, correctID := range correctList {
			if idStr == correctID {
				isCorrect = true
				break
			}
		}
		subScore := 0.0
		if isCorrect {
			subScore = perAnswerScore
			score += subScore
		} else {
			isAllCorrect = false
		}
		records = append(records, &models.ExamQuestionUser{
			ExamID:     req.ExamId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			AnswerID:   id,
			IsCorrect:  isCorrect,
			Score:      subScore,
		})
	}
	if err := s.repo.SaveBatchExamQuestionUserMultipleChoice(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	// Chuẩn bị response
	trueAnswerIDs := make([]int64, len(correctList))
	trueAnswerContents := make([]string, len(correctList))
	userAnswerIDs := make([]int64, len(req.AnswerIds))
	userAnswerContents := make([]string, len(req.AnswerIds))
	for i, id := range correctList {
		trueAnswerIDs[i], _ = strconv.ParseInt(id, 10, 64)
		trueAnswerContents[i] = idToContent[id]
	}
	for i, id := range req.AnswerIds {
		userAnswerIDs[i] = id
		userAnswerContents[i] = idToContent[strconv.FormatInt(id, 10)]
	}
	return &prot.SaveScoreResponse{
		Id:            0,
		QuestionId:    req.QuestionId,
		AnswerIds:     userAnswerIDs,
		TrueAnswerIds: trueAnswerIDs,
		Answers:       userAnswerContents,
		TrueAnswers:   trueAnswerContents,
		IsCorrect:     isAllCorrect,
		Score:         utils.RoundTo2Decimal(score),
	}, nil
}


func (s *saveScoreMultipleChoiceService) SaveScoreMultipleChoiceHomework(req *prot.SaveScoreMultipleChoiceRequest, userID int64) (*prot.SaveScoreResponse, error) {
	if req.HomeworkId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.HomeworkId, "homework")
	if err != nil {
		return nil, err
	}
	q, ok := questions[strconv.FormatInt(req.QuestionId, 10)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	var raw struct {
		Ids  []interface{}     `json:"ids"`
		List map[string]string `json:"list"`
	}

	err = json.Unmarshal(q.CorrectAnswers, &raw)
	if err != nil {
		return nil, fmt.Errorf("Dữ liệu correct_answers không đúng định dạng: %v", err)
	}
	correctAnswers := struct {
		Ids  []string
		List map[string]string
	}{
		List: raw.List,
	}

	for _, v := range raw.Ids {
		switch val := v.(type) {
		case float64:
			correctAnswers.Ids = append(correctAnswers.Ids, fmt.Sprintf("%.0f", val))
		case string:
			correctAnswers.Ids = append(correctAnswers.Ids, val)
		default:
			correctAnswers.Ids = append(correctAnswers.Ids, fmt.Sprint(val))
		}
	}

	var correctList []string
	if len(correctAnswers.Ids) > 0 {
		correctList = correctAnswers.Ids
	} else if len(correctAnswers.List) > 0 {
		for k := range correctAnswers.List {
			correctList = append(correctList, k)
		}
	} else {
		return nil, fmt.Errorf("Không tìm thấy đáp án đúng trong correct_answers")
	}
	// Parse options.answers để lấy map answer_id -> content
	var options struct {
		Answers []struct {
			ID      json.Number `json:"id"`
			Content string      `json:"text"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	idToContent := make(map[string]string)
	for _, ans := range options.Answers {
		idToContent[ans.ID.String()] = ans.Content
	}

	perAnswerScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Kiểm tra và xóa câu trả lời cũ nếu cần
	needSave, err := CheckAndCleanExistingAnswersWithTx(tx, "homework_question_users", req.HomeworkId, userID, req.QuestionId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Tính toán kết quả để trả về response
	var records []*models.HomeworkQuestionUser
	isAllCorrect := true
	score := 0.0
	for _, id := range req.AnswerIds {
		isCorrect := false
		idStr := strconv.FormatInt(id, 10)
		for _, correctID := range correctList {
			if idStr == correctID {
				isCorrect = true
				break
			}
		}
		subScore := 0.0
		if isCorrect {
			subScore = perAnswerScore
			score += subScore
		} else {
			isAllCorrect = false
		}
		records = append(records, &models.HomeworkQuestionUser{
			HomeworkID: req.HomeworkId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			AnswerID:   id,
			IsCorrect:  isCorrect,
			Score:      subScore,
		})
	}

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

	// Nếu không cần lưu (đã có câu trả lời đúng hết), query từ DB để lấy star, ratio_score, weight, number_time_sent
	if !needSave {
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
		// Chuẩn bị response
		trueAnswerIDs := make([]int64, len(correctList))
		trueAnswerContents := make([]string, len(correctList))
		userAnswerIDs := make([]int64, len(req.AnswerIds))
		userAnswerContents := make([]string, len(req.AnswerIds))
		for i, id := range correctList {
			trueAnswerIDs[i], _ = strconv.ParseInt(id, 10, 64)
			trueAnswerContents[i] = idToContent[id]
		}
		for i, id := range req.AnswerIds {
			userAnswerIDs[i] = id
			userAnswerContents[i] = idToContent[strconv.FormatInt(id, 10)]
		}
		return &prot.SaveScoreResponse{
			QuestionId:      req.QuestionId,
			AnswerIds:       userAnswerIDs,
			TrueAnswerIds:   trueAnswerIDs,
			Answers:         userAnswerContents,
			TrueAnswers:     trueAnswerContents,
			IsCorrect:       isAllCorrect,
			Score:           utils.RoundTo2Decimal(score),
			Star:            int32(star),
			RatioScore:      ratioScore,
			Weight:          weight,
			NumberTimeSent:  int32(numberTimeSent),
		}, nil
	}

	// Upsert trạng thái hoàn thành nếu đúng hết
	if isAllCorrect {
		err := s.correctRepo.UpsertHomeworkUserOnCorrect(req.HomeworkId, req.LessonId, userID, req.QuestionId, "multiple_choice")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// Lưu kết quả vào DB
	if err := s.repo.SaveBatchHomeworkQuestionUserMultipleChoice(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	// Chuẩn bị response
	trueAnswerIDs := make([]int64, len(correctList))
	trueAnswerContents := make([]string, len(correctList))
	userAnswerIDs := make([]int64, len(req.AnswerIds))
	userAnswerContents := make([]string, len(req.AnswerIds))
	for i, id := range correctList {
		trueAnswerIDs[i], _ = strconv.ParseInt(id, 10, 64)
		trueAnswerContents[i] = idToContent[id]
	}
	for i, id := range req.AnswerIds {
		userAnswerIDs[i] = id
		userAnswerContents[i] = idToContent[strconv.FormatInt(id, 10)]
	}
	return &prot.SaveScoreResponse{
		Id:             0,
		QuestionId:     req.QuestionId,
		AnswerIds:      userAnswerIDs,
		TrueAnswerIds:  trueAnswerIDs,
		Answers:        userAnswerContents,
		TrueAnswers:    trueAnswerContents,
		IsCorrect:      isAllCorrect,
		Score:          utils.RoundTo2Decimal(score),
		Star:           int32(star),
		RatioScore:     ratioScore,
		Weight:         weight,
		NumberTimeSent: int32(numberTimeSent),
	}, nil
}

func (s *saveScoreMultipleChoiceService) SaveScoreMultipleChoiceLevelTest(req *prot.SaveScoreMultipleChoiceRequest, userID int64) (*prot.SaveScoreResponse, error) {
	if req.LevelTestId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	trueAnswers, userAnswers, isCorrect, score, err := s.prepareScoreData(req)
	if err != nil {
		return nil, err
	}

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lưu từng đáp án của user
	for _, userAnswer := range userAnswers {
		record := &models.LevelTestQuestionUser{
			LevelTestID:  req.LevelTestId,
			UserID:       userID,
			QuestionID:   req.QuestionId,
			AnswerID:     userAnswer.ID,
			TrueAnswerID: 0, // Không cần lưu true_answer_id nữa vì có thể có nhiều đáp án đúng
			IsCorrect:    isCorrect,
			Score:        score,
		}

		if err := s.repo.SaveLevelTestQuestionUserMultipleChoice(record, tx); err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	tx.Commit()

	// Chuẩn bị response
	trueAnswerIDs := make([]int64, len(trueAnswers))
	trueAnswerContents := make([]string, len(trueAnswers))
	userAnswerIDs := make([]int64, len(userAnswers))
	userAnswerContents := make([]string, len(userAnswers))

	for i, ans := range trueAnswers {
		trueAnswerIDs[i] = ans.ID
		trueAnswerContents[i] = ans.Content
	}

	for i, ans := range userAnswers {
		userAnswerIDs[i] = ans.ID
		userAnswerContents[i] = ans.Content
	}

	return &prot.SaveScoreResponse{
		Id:            0, // Không cần ID nữa vì có thể có nhiều bản ghi
		QuestionId:    req.QuestionId,
		AnswerIds:     userAnswerIDs,
		TrueAnswerIds: trueAnswerIDs,
		Answers:       userAnswerContents,
		TrueAnswers:   trueAnswerContents,
		IsCorrect:     isCorrect,
		Score:         utils.RoundTo2Decimal(score),
	}, nil
}

func (s *saveScoreMultipleChoiceService) SaveScoreMultipleChoiceExercise(req *prot.SaveScoreMultipleChoiceRequest, userID int64) (*prot.SaveScoreResponse, error) {
	if req.ExerciseId == 0 {
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

	var raw struct {
		Ids  []interface{}     `json:"ids"`
		List map[string]string `json:"list"`
	}

	err = json.Unmarshal(q.CorrectAnswers, &raw)
	if err != nil {
		return nil, fmt.Errorf("Dữ liệu correct_answers không đúng định dạng: %v", err)
	}
	correctAnswers := struct {
		Ids  []string
		List map[string]string
	}{
		List: raw.List,
	}

	for _, v := range raw.Ids {
		switch val := v.(type) {
		case float64:
			correctAnswers.Ids = append(correctAnswers.Ids, fmt.Sprintf("%.0f", val))
		case string:
			correctAnswers.Ids = append(correctAnswers.Ids, val)
		default:
			correctAnswers.Ids = append(correctAnswers.Ids, fmt.Sprint(val))
		}
	}

	var correctList []string
	if len(correctAnswers.Ids) > 0 {
		correctList = correctAnswers.Ids
	} else if len(correctAnswers.List) > 0 {
		for k := range correctAnswers.List {
			correctList = append(correctList, k)
		}
	} else {
		return nil, fmt.Errorf("Không tìm thấy đáp án đúng trong correct_answers")
	}
	// Parse options.answers để lấy map answer_id -> content
	var options struct {
		Answers []struct {
			ID      json.Number `json:"id"`
			Content string      `json:"text"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.answers: %v", err)
	}
	idToContent := make(map[string]string)
	for _, ans := range options.Answers {
		idToContent[ans.ID.String()] = ans.Content
	}

	perAnswerScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var records []*models.ExerciseQuestionUser
	isAllCorrect := true
	score := 0.0
	for _, id := range req.AnswerIds {
		isCorrect := false
		idStr := strconv.FormatInt(id, 10)
		for _, correctID := range correctList {
			if idStr == correctID {
				isCorrect = true
				break
			}
		}
		subScore := 0.0
		if isCorrect {
			subScore = perAnswerScore
			score += subScore
		} else {
			isAllCorrect = false
		}
		records = append(records, &models.ExerciseQuestionUser{
			ExerciseID: req.ExerciseId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			AnswerID:   id,
			IsCorrect:  isCorrect,
			Score:      subScore,
		})
	}
	if err := s.repo.SaveBatchExerciseQuestionUserMultipleChoice(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	// Chuẩn bị response
	trueAnswerIDs := make([]int64, len(correctList))
	trueAnswerContents := make([]string, len(correctList))
	userAnswerIDs := make([]int64, len(req.AnswerIds))
	userAnswerContents := make([]string, len(req.AnswerIds))
	for i, id := range correctList {
		trueAnswerIDs[i], _ = strconv.ParseInt(id, 10, 64)
		trueAnswerContents[i] = idToContent[id]
	}
	for i, id := range req.AnswerIds {
		userAnswerIDs[i] = id
		userAnswerContents[i] = idToContent[strconv.FormatInt(id, 10)]
	}
	return &prot.SaveScoreResponse{
		Id:            0,
		QuestionId:    req.QuestionId,
		AnswerIds:     userAnswerIDs,
		TrueAnswerIds: trueAnswerIDs,
		Answers:       userAnswerContents,
		TrueAnswers:   trueAnswerContents,
		IsCorrect:     isAllCorrect,
		Score:         utils.RoundTo2Decimal(score),
	}, nil
}

