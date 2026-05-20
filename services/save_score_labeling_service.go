package services

import (
	"be-cleverschool/database/db"
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"encoding/json"
	"fmt"
	"strconv"
)

type SaveScoreLabelingService interface {
	SaveScoreLabelingExam(req *prot.SaveScoreLabelingRequest, userID int64) (*prot.SaveScoreResponseLabeling, error)
	SaveScoreLabelingHomework(req *prot.SaveScoreLabelingRequest, userID int64) (*prot.SaveScoreResponseLabeling, error)
	SaveScoreLabelingExercise(req *prot.SaveScoreLabelingRequest, userID int64) (*prot.SaveScoreResponseLabeling, error)
	//SaveScoreLabelingLevelTest(req *requests.SaveScoreLabelingRequest, userID int64) (*prot.SaveScoreResponseLabeling, error)
}

type saveScoreLabelingService struct {
	repo                        repositories.SaveScoreLabelingRepository
	correctRepo                 repositories.SaveCorrectHomeworkRepository
	clonedQuestionService       ClonedQuestionService
	homeworkUserQuestionService HomeworkUserQuestionService
}

func NewSaveScoreLabelingService(repo repositories.SaveScoreLabelingRepository, clonedQuestionService ClonedQuestionService) SaveScoreLabelingService {
	homeworkUserQuestionRepo := repositories.NewHomeworkUserQuestionRepository()
	return &saveScoreLabelingService{
		repo:                        repo,
		correctRepo:                 repositories.NewSaveCorrectHomeworkRepository(),
		clonedQuestionService:       clonedQuestionService,
		homeworkUserQuestionService: NewHomeworkUserQuestionService(homeworkUserQuestionRepo, clonedQuestionService),
	}
}

func (s *saveScoreLabelingService) SaveScoreLabelingExam(req *prot.SaveScoreLabelingRequest, userID int64) (*prot.SaveScoreResponseLabeling, error) {
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

	// Parse options.labels để lấy map label_id -> group_position, text
	var options struct {
		Labels []struct {
			ID            json.Number `json:"id"`
			GroupPosition int         `json:"group_position"`
			Text          string      `json:"text"`
		} `json:"labels"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.labels: %v", err)
	}
	labelIdToGroupPos := make(map[int64]int)
	labelIdToText := make(map[int64]string)
	for _, l := range options.Labels {
		if id, err := strconv.ParseInt(l.ID.String(), 10, 64); err == nil {
			labelIdToGroupPos[id] = l.GroupPosition
			labelIdToText[id] = l.Text
		}
	}

	numLabels := len(req.Answers)
	perLabelScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var labelResults []*prot.LabelingPair
	var records []*models.ExamQuestionUserLabeling
	for blankIDStr, labelID := range req.Answers {
		var blankID int64
		if blankId, err := strconv.ParseInt(blankIDStr, 10, 64); err == nil {
			blankID = blankId
		}
		blankGroupPos := labelIdToGroupPos[blankID]
		labelGroupPos := labelIdToGroupPos[labelID]
		isCorrect := blankGroupPos == labelGroupPos
		score := 0.0
		if isCorrect {
			score = perLabelScore
			correctCount++
			totalScore += score
		}
		labelResults = append(labelResults, &prot.LabelingPair{
			BlankId:       blankID,
			AnswerId:      labelID,
			BlankContent:  labelIdToText[blankID],
			AnswerContent: labelIdToText[labelID],
			IsCorrect:     isCorrect,
			Score:         score,
		})
		records = append(records, &models.ExamQuestionUserLabeling{
			ExamID:     req.ExamId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			BlankID:    blankID,
			AnswerID:   labelID,
			IsCorrect:  isCorrect,
			Score:      score,
		})
	}
	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := s.repo.SaveBatchExamQuestionUserLabeling(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseLabeling{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Labels:       labelResults,
		IsAllCorrect: correctCount == numLabels,
	}, nil
}

func (s *saveScoreLabelingService) SaveScoreLabelingHomework(req *prot.SaveScoreLabelingRequest, userID int64) (*prot.SaveScoreResponseLabeling, error) {
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

	// Parse options.labels để lấy map label_id -> group_position, text
	var options struct {
		Labels []struct {
			ID            json.Number `json:"id"`
			GroupPosition int         `json:"group_position"`
			Text          string      `json:"text"`
		} `json:"labels"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.labels: %v", err)
	}
	labelIdToGroupPos := make(map[int64]int)
	labelIdToText := make(map[int64]string)
	for _, l := range options.Labels {
		if id, err := strconv.ParseInt(l.ID.String(), 10, 64); err == nil {
			labelIdToGroupPos[id] = l.GroupPosition
			labelIdToText[id] = l.Text
		}
	}

	numLabels := len(req.Answers)
	perLabelScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var labelResults []*prot.LabelingPair
	var records []*models.HomeworkQuestionUserLabeling
	for blankIDStr, labelID := range req.Answers {
		var blankID int64
		if blankId, err := strconv.ParseInt(blankIDStr, 10, 64); err == nil {
			blankID = blankId
		}
		blankGroupPos := labelIdToGroupPos[blankID]
		labelGroupPos := labelIdToGroupPos[labelID]
		isCorrect := blankGroupPos == labelGroupPos
		score := 0.0
		if isCorrect {
			score = perLabelScore
			correctCount++
			totalScore += score
		}
		labelResults = append(labelResults, &prot.LabelingPair{
			BlankId:       blankID,
			AnswerId:      labelID,
			BlankContent:  labelIdToText[blankID],
			AnswerContent: labelIdToText[labelID],
			IsCorrect:     isCorrect,
			Score:         score,
		})
		records = append(records, &models.HomeworkQuestionUserLabeling{
			HomeworkID: req.HomeworkId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			BlankID:    blankID,
			AnswerID:   labelID,
			IsCorrect:  isCorrect,
			Score:      score,
		})
	}
	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Kiểm tra và xóa câu trả lời cũ nếu cần
	needSave, err := CheckAndCleanExistingAnswersWithTx(tx, "homework_question_user_labelings", req.HomeworkId, userID, req.QuestionId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	isAllCorrect := correctCount == numLabels

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
		return &prot.SaveScoreResponseLabeling{
			QuestionId:      req.QuestionId,
			TotalScore:      utils.RoundTo2Decimal(totalScore),
			Labels:          labelResults,
			IsAllCorrect:    isAllCorrect,
			Star:            int32(star),
			RatioScore:      ratioScore,
			Weight:          weight,
			NumberTimeSent:  int32(numberTimeSent),
		}, nil
	}

	// Upsert trạng thái hoàn thành nếu đúng hết
	if isAllCorrect {
		err := s.correctRepo.UpsertHomeworkUserOnCorrect(req.HomeworkId, req.LessonId, userID, req.QuestionId, "labeling")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	// Lưu dữ liệu hàng loạt
	if err := s.repo.SaveBatchHomeworkQuestionUserLabeling(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &prot.SaveScoreResponseLabeling{
		QuestionId:      req.QuestionId,
		TotalScore:      utils.RoundTo2Decimal(totalScore),
		Labels:          labelResults,
		IsAllCorrect:    isAllCorrect,
		Star:            int32(star),
		RatioScore:      ratioScore,
		Weight:          weight,
		NumberTimeSent:  int32(numberTimeSent),
	}, nil
}

func (s *saveScoreLabelingService) SaveScoreLabelingExercise(req *prot.SaveScoreLabelingRequest, userID int64) (*prot.SaveScoreResponseLabeling, error) {
	if req.ExerciseId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	// Lấy danh sách câu hỏi từ cloned_question
	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExerciseId, "exercise")
	if err != nil {
		return nil, err
	}
	q, ok := questions[fmt.Sprintf("%v", req.QuestionId)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	// Parse options.labels để lấy map label_id -> group_position, text
	var options struct {
		Labels []struct {
			ID            json.Number `json:"id"`
			GroupPosition int         `json:"group_position"`
			Text          string      `json:"text"`
		} `json:"labels"`
	}
	if err := json.Unmarshal(q.Options, &options); err != nil {
		return nil, fmt.Errorf("Lỗi parse options.labels: %v", err)
	}
	labelIdToGroupPos := make(map[int64]int)
	labelIdToText := make(map[int64]string)
	for _, l := range options.Labels {
		if id, err := strconv.ParseInt(l.ID.String(), 10, 64); err == nil {
			labelIdToGroupPos[id] = l.GroupPosition
			labelIdToText[id] = l.Text
		}
	}

	numLabels := len(req.Answers)
	perLabelScore := 1.0 // hoặc lấy từ q.Metadata nếu có
	totalScore := 0.0
	correctCount := 0
	var labelResults []*prot.LabelingPair
	var records []*models.ExerciseQuestionUserLabeling
	for blankIDStr, labelID := range req.Answers {
		var blankID int64
		if blankId, err := strconv.ParseInt(blankIDStr, 10, 64); err == nil {
			blankID = blankId
		}
		blankGroupPos := labelIdToGroupPos[blankID]
		labelGroupPos := labelIdToGroupPos[labelID]
		isCorrect := blankGroupPos == labelGroupPos
		score := 0.0
		if isCorrect {
			score = perLabelScore
			correctCount++
			totalScore += score
		}
		labelResults = append(labelResults, &prot.LabelingPair{
			BlankId:       blankID,
			AnswerId:      labelID,
			BlankContent:  labelIdToText[blankID],
			AnswerContent: labelIdToText[labelID],
			IsCorrect:     isCorrect,
			Score:         score,
		})
		records = append(records, &models.ExerciseQuestionUserLabeling{
			ExerciseID: req.ExerciseId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			BlankID:    blankID,
			AnswerID:   labelID,
			IsCorrect:  isCorrect,
			Score:      score,
		})
	}
	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := s.repo.SaveBatchExerciseQuestionUserLabeling(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseLabeling{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Labels:       labelResults,
		IsAllCorrect: correctCount == numLabels,
	}, nil
}

//func (s *saveScoreLabelingService) SaveScoreLabelingLevelTest(req *requests.SaveScoreLabelingRequest, userID int64) (*prot.SaveScoreResponseLabeling, error) {
//	if req.LevelTestID == 0 || len(req.Answers) == 0 {
//		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
//	}
//
//	// Lấy tổng điểm của câu hỏi để chia đều cho mỗi cặp labeling
//	questionScore, err := s.repo.GetLevelTestQuestionScoreLabeling(req.LevelTestID, req.QuestionID)
//	if err != nil {
//		return nil, err
//	}
//
//	numLabels := len(req.Answers)
//	perLabelScore := questionScore / float64(numLabels)
//	perLabelScore = math.Round(perLabelScore*100) / 100 // Làm tròn đến 2 chữ số sau dấu phẩy
//
//	tx := db.MasterDB.Begin()
//	defer func() {
//		if r := recover(); r != nil {
//			tx.Rollback()
//		}
//	}()
//
//	var (
//		records      []*models.LevelTestQuestionUserLabeling
//		labelResults []*prot.LabelingPair
//		totalScore   float64
//		correctCount int
//	)
//
//	// Xử lý từng cặp labeling của người dùng
//	for blankID, answerID := range req.Answers {
//		// Lấy thông tin của blank và answer
//		blank, err := s.repo.GetAnswerByID(blankID)
//		if err != nil {
//			tx.Rollback()
//			return nil, err
//		}
//
//		answer, err := s.repo.GetAnswerByID(answerID)
//		if err != nil {
//			tx.Rollback()
//			return nil, err
//		}
//
//		// Kiểm tra labeling đúng
//		isCorrect := blankID == answerID
//		score := 0.0
//		if isCorrect {
//			score = perLabelScore
//			correctCount++
//			totalScore += score
//		}
//
//		record := &models.LevelTestQuestionUserLabeling{
//			LevelTestID: req.LevelTestID,
//			UserID:      userID,
//			QuestionID:  req.QuestionID,
//			BlankID:     blankID,
//			AnswerID:    answerID,
//			IsCorrect:   isCorrect,
//			Score:       score,
//		}
//		records = append(records, record)
//
//		labelResults = append(labelResults, &prot.LabelingPair{
//			BlankId:       blankID,
//			AnswerId:      answerID,
//			BlankContent:  blank.Content,
//			AnswerContent: answer.MatchingContent,
//			IsCorrect:     isCorrect,
//			Score:         score,
//		})
//	}
//
//	// Lưu dữ liệu hàng loạt
//	if err := s.repo.SaveBatchLevelTestQuestionUserLabeling(records, tx); err != nil {
//		tx.Rollback()
//		return nil, err
//	}
//	tx.Commit()
//
//	return &prot.SaveScoreResponseLabeling{
//		QuestionId:   req.QuestionID,
//		TotalScore:   utils.RoundTo2Decimal(totalScore),
//		Labels:       labelResults,
//		IsAllCorrect: correctCount == numLabels,
//	}, nil
//}

