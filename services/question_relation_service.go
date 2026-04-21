package services

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type QuestionRelationService interface {
	AssignQuestions(c *gin.Context, req *prot.CreateQuestionRelationRequest) error
	CreateOrUpdateCloneQuestion(c *gin.Context, assignmentId int64, assignmentType string, filterKey string, question *models.Question, req *prot.CreateQuestionRelationRequest, clonedQuestion *models.ClonedQuestion) error
}

type questionRelationService struct {
	repo repositories.QuestionRelationRepository
}

func NewQuestionRelationService(repo repositories.QuestionRelationRepository) QuestionRelationService {
	return &questionRelationService{repo: repo}
}

func (s *questionRelationService) AssignQuestions(c *gin.Context, req *prot.CreateQuestionRelationRequest) error {
	hasOrderData := len(req.QuestionsOrder) > 0 || len(req.SourceQuestionsOrder) > 0
	emptyScoresAndFlags := len(req.QuestionScores) == 0 && len(req.SourceQuestionScores) == 0 && !req.IsReset && !req.ClearSelfScoringQuestions

	if emptyScoresAndFlags && !hasOrderData {
		// Xóa theo exercise nếu có
		if req.ExerciseId > 0 {
			return s.repo.DeleteAllExerciseQuestions(int64(req.ExerciseId), db.MasterDB)
		}
		// Chỉ xóa khi có exam_id hoặc homework_id thực tế
		if req.ExamId > 0 || req.HomeworkId > 0 {
			return s.repo.DeleteAllQuestionsInAssignment(int64(req.ExamId), int64(req.HomeworkId))
		}
		// Nếu không có exam_id, homework_id, contest_round_id, level_test_id, lesson_plan_part_id, không làm gì cả
		if req.ContestRoundId == 0 && req.LevelTestId == 0 && req.LessonPlanPartId == 0 {
			return nil
		}
	}

	// Skip check assigned by homework
	// isAssignedHomework, _ := s.repo.IsAssignedHomework(int64(req.HomeworkID), 0)
	isAssignedHomework := false
	isAssignedExam, _ := s.repo.IsAssignedExam(int64(req.ExamId), 0)
	// isAssignedExercise removed: always update exercise cloned_questions

	var err error

	// err = db.MasterDB.Transaction(func(tx *gorm.DB) error {
	// 	// Xử lý câu hỏi chính
	// 	if len(req.QuestionScores) > 0 {
	// 		if req.ExamID > 0 && !isAssignedHomework {
	// 			if err := s.repo.BulkInsertExamQuestions(req.ExamID, req.QuestionScores, tx); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		if req.LessonPlanPartID > 0 {
	// 			if err := s.repo.BulkInsertLessonPlanPartQuestions(req.LessonPlanPartID, req.QuestionScores, tx); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		if req.HomeworkID > 0 && !isAssignedExam {
	// 			if err := s.repo.BulkInsertHomeworkQuestions(req.HomeworkID, req.QuestionScores, tx); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		if req.LevelTestID > 0 {
	// 			if err := s.repo.BulkInsertLevelTestQuestions(req.LevelTestID, req.QuestionScores, tx); err != nil {
	// 				return err
	// 			}
	// 		}
	// 	}

	// 	// Xử lý câu hỏi source (ví dụ: bản nháp hoặc backup)
	// 	if len(req.SourceQuestionScores) > 0 {
	// 		if req.ExamID > 0 && !isAssignedExam {
	// 			if err := s.repo.BulkInsertExamSourceQuestions(req.ExamID, req.SourceQuestionScores, tx); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		if req.LessonPlanPartID > 0 {
	// 			if err := s.repo.BulkInsertLessonPlanPartSourceQuestions(req.LessonPlanPartID, req.SourceQuestionScores, tx); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		if req.HomeworkID > 0 && !isAssignedHomework {
	// 			if err := s.repo.BulkInsertHomeworkSourceQuestions(req.HomeworkID, req.SourceQuestionScores, tx); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		if req.LevelTestID > 0 {
	// 			if err := s.repo.BulkInsertLevelTestSourceQuestions(req.LevelTestID, req.SourceQuestionScores, tx); err != nil {
	// 				return err
	// 			}
	// 		}
	// 	}

	// 	return nil
	// })

	var exams []models.Exam
	var homeworks []models.Homework

	if req.ResetAll {
		// If specific assignment ID is provided, only reset that one
		if req.ContestRoundId > 0 {
			return s.UpdateCloneQuestion(c.Copy(), int64(req.ContestRoundId), models.ClonedQuestionTypeContestRound, req)
		}
		if req.HomeworkId > 0 {
			return s.UpdateCloneQuestion(c.Copy(), int64(req.HomeworkId), models.ClonedQuestionTypeHomework, req)
		}
		if req.ExamId > 0 {
			return s.UpdateCloneQuestion(c.Copy(), int64(req.ExamId), models.ClonedQuestionTypeExam, req)
		}
		if req.LevelTestId > 0 {
			return s.UpdateCloneQuestion(c.Copy(), int64(req.LevelTestId), models.ClonedQuestionTypeLevelTest, req)
		}
		if req.LessonPlanPartId > 0 {
			return s.UpdateCloneQuestion(c.Copy(), int64(req.LessonPlanPartId), models.ClonedQuestionTypeLessonPlanPart, req)
		}
		if req.ExerciseId > 0 {
			return s.UpdateCloneQuestion(c.Copy(), int64(req.ExerciseId), models.ClonedQuestionTypeExercise, req)
		}

		// If no specific assignment ID, reset all (original behavior)
		examRepo := repositories.NewExamRepository()
		homeworkRepo := repositories.NewHomeworkRepository()

		homeworks, err = homeworkRepo.GetAll()
		if err != nil {
			return err
		}

		for _, homework := range homeworks {
			err := s.UpdateCloneQuestion(c.Copy(), int64(homework.ID), models.ClonedQuestionTypeHomework, req)
			if err != nil {
				return err
			}
		}

		exams, err = examRepo.GetAll()
		if err != nil {
			return err
		}

		for _, exam := range exams {
			err := s.UpdateCloneQuestion(c.Copy(), int64(exam.ID), models.ClonedQuestionTypeExam, req)
			if err != nil {
				return err
			}
		}
		return nil
	} else if req.ClearSelfScoringQuestions {
		homeworkRepo := repositories.NewHomeworkRepository()
		homeworks, err = homeworkRepo.GetAll()

		for _, homework := range homeworks {
			err := s.UpdateCloneQuestion(c.Copy(), int64(homework.ID), models.ClonedQuestionTypeHomework, req)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if err == nil {
		useOrderMode := len(req.QuestionsOrder) > 0 || len(req.SourceQuestionsOrder) > 0
		if len(req.QuestionScores) > 0 || len(req.SourceQuestionScores) > 0 || req.IsReset || useOrderMode {
			if req.ExamId > 0 {
				if !isAssignedExam || req.IsReset {
					return s.UpdateCloneQuestion(c.Copy(), int64(req.ExamId), models.ClonedQuestionTypeExam, req)
				} else {
					return fmt.Errorf(i18n.Localize("messages.error_is_assigned"))
				}
			}
			if req.ExerciseId > 0 {
				return s.UpdateCloneQuestion(c.Copy(), int64(req.ExerciseId), models.ClonedQuestionTypeExercise, req)
			}
			if req.LessonPlanPartId > 0 {
				return s.UpdateCloneQuestion(c.Copy(), int64(req.LessonPlanPartId), models.ClonedQuestionTypeLessonPlanPart, req)
			}
			if req.HomeworkId > 0 {
				// Skip check assigned by homework
				if !isAssignedHomework || req.IsReset {
					return s.UpdateCloneQuestion(c.Copy(), int64(req.HomeworkId), models.ClonedQuestionTypeHomework, req)
				} else {
					return fmt.Errorf(i18n.Localize("messages.error_is_assigned"))
				}
			}
			if req.LevelTestId > 0 {
				return s.UpdateCloneQuestion(c.Copy(), int64(req.LevelTestId), models.ClonedQuestionTypeLevelTest, req)
			}
			if req.ContestRoundId > 0 {
				return s.UpdateCloneQuestion(c.Copy(), int64(req.ContestRoundId), models.ClonedQuestionTypeContestRound, req)
			}
		}
	}

	return err
}

func (s *questionRelationService) UpdateCloneQuestion(c *gin.Context, assignmentId int64, assignmentType string, req *prot.CreateQuestionRelationRequest) error {
	clonedRepo := repositories.NewClonedQuestionRepository()
	questionRepo := repositories.NewQuestionRepository()
	questionService := NewQuestionService(questionRepo)
	clonedData, err := clonedRepo.FindByAssignment(assignmentId, assignmentType)
	filterKey, _ := questionService.GetKey(assignmentType)

	if errors.Is(err, gorm.ErrRecordNotFound) || clonedData == nil || clonedData.ID == 0 {
		return s.CreateOrUpdateCloneQuestion(c, assignmentId, assignmentType, filterKey, &models.Question{}, req, &models.ClonedQuestion{})
	} else {
		return s.CreateOrUpdateCloneQuestion(c, assignmentId, assignmentType, filterKey, &models.Question{}, req, clonedData)
	}
}

func (s *questionRelationService) CreateOrUpdateCloneQuestion(c *gin.Context, assignmentId int64, assignmentType string, filterKey string, question *models.Question, req *prot.CreateQuestionRelationRequest, clonedQuestion *models.ClonedQuestion) error {
	questionResource := resources.NewQuestionResource()
	clonedRepo := repositories.NewClonedQuestionRepository()
	clonedRepo.SetContext(c)

	questionRepo := repositories.NewQuestionRepository()

	questionIds := make([]int, 0, len(req.QuestionScores))
	sourceQuestionIds := make([]int, 0, len(req.SourceQuestionScores))

	if clonedQuestion.ID != 0 {
		if req.IsReset {
			var existingRawQuestions []json.RawMessage
			if clonedQuestion.Questions != nil {
				if err := json.Unmarshal(clonedQuestion.Questions, &existingRawQuestions); err != nil {
					config.Log.Error("unmarshal clonedQuestion.Questions error:", err.Error())
					return err
				}
			}

			for _, q := range existingRawQuestions {
				question, err := questionResource.ParseProtQuestionFromJSON(q)
				if err != nil {
					config.Log.Error("parse question error:", err.Error())
					return nil
				}

				questionIds = append(questionIds, int(question.Id))
			}
		} else if req.ClearSelfScoringQuestions {
			var existingRawQuestions []json.RawMessage
			if clonedQuestion.Questions != nil {
				if err := json.Unmarshal(clonedQuestion.Questions, &existingRawQuestions); err != nil {
					config.Log.Error("unmarshal clonedQuestion.Questions error:", err.Error())
					return err
				}
			}

			list := make([]*prot.Question, 0, len(existingRawQuestions))

			for _, q := range existingRawQuestions {
				question, err := questionResource.ParseProtQuestionFromJSON(q)
				if err != nil {
					config.Log.Error("parse question error:", err.Error())
					return nil
				}

				if question.Type != models.QuestionTypeSpeaking && question.Type != models.QuestionTypeWriting {
					formatted := questionResource.FormatStripDomain(question)
					list = append(list, formatted)
				}
			}

			marshalOptions := protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseProtoNames:   true,
			}

			var buf strings.Builder
			buf.WriteString("[")
			for i, q := range list {
				b, err := marshalOptions.Marshal(q)
				if err != nil {
					config.Log.Errorf("marshal question failed: %v", err)
					continue
				}
				buf.Write(b)
				if i != len(list)-1 {
					buf.WriteString(",")
				}
			}
			buf.WriteString("]")

			clonedQuestion.Questions = []byte(buf.String())
			clonedRepo.SetContext(c)
			clonedRepo.Update(clonedQuestion)

			if assignmentType == models.ClonedQuestionTypeExam {
				_ = UpdateExamTotalQuestions(assignmentId)
			} else if assignmentType == models.ClonedQuestionTypeHomework {
				_ = UpdateHomeworkTotalQuestions(assignmentId)
			} else if assignmentType == models.ClonedQuestionTypeExercise {
				_ = UpdateExerciseTotalQuestions(assignmentId)
			}

			return nil
		} else {
			for idStr := range req.QuestionScores {
				var id int
				if idNew, err := strconv.Atoi(idStr); err == nil {
					id = idNew
				}
				questionIds = append(questionIds, id)
			}
		}
		for idStr := range req.SourceQuestionScores {
			var id int
			if idNew, err := strconv.Atoi(idStr); err == nil {
				id = idNew
			}
			sourceQuestionIds = append(sourceQuestionIds, id)
		}
	} else {
		for idStr := range req.QuestionScores {
			var id int
			if idNew, err := strconv.Atoi(idStr); err == nil {
				id = idNew
			}
			questionIds = append(questionIds, id)
		}
		for idStr := range req.SourceQuestionScores {
			var id int
			if idNew, err := strconv.Atoi(idStr); err == nil {
				id = idNew
			}
			sourceQuestionIds = append(sourceQuestionIds, id)
		}
	}

	// Also include IDs from questions_order / source_questions_order
	useOrder := len(req.QuestionsOrder) > 0 || len(req.SourceQuestionsOrder) > 0
	if useOrder {
		for idStr := range req.QuestionsOrder {
			if idNew, err := strconv.Atoi(idStr); err == nil {
				questionIds = append(questionIds, idNew)
			}
		}
		for idStr := range req.SourceQuestionsOrder {
			if idNew, err := strconv.Atoi(idStr); err == nil {
				sourceQuestionIds = append(sourceQuestionIds, idNew)
			}
		}
	}

	// Build comma-separated list of source question IDs (in the same order as processed)
	var sortSourceQuestionIds string
	if len(sourceQuestionIds) > 0 {
		sourceIdStrs := make([]string, len(sourceQuestionIds))
		for i, id := range sourceQuestionIds {
			sourceIdStrs[i] = strconv.Itoa(id)
		}
		sortSourceQuestionIds = strings.Join(sourceIdStrs, ",")
	}

	idStrs := make([]string, len(questionIds))
	idSet := make(map[string]bool)

	for i, id := range questionIds {
		idStr := strconv.Itoa(id)
		idStrs[i] = idStr
		idSet[idStr] = true
	}

	if len(sourceQuestionIds) > 0 {
		sourceIdStrs := make([]string, len(sourceQuestionIds))
		for i, id := range sourceQuestionIds {
			sourceIdStrs[i] = strconv.Itoa(id)
		}
		questionRepo.SetFilter(map[string]interface{}{
			"source_question_id": "in:" + strings.Join(sourceIdStrs, ","),
		})
		questionBySourceIds, err := questionRepo.GetAll()
		if err != nil {
			return err
		}

		for _, q := range questionBySourceIds {
			idStr := strconv.Itoa(int(q.ID))
			if !idSet[idStr] {
				idStrs = append(idStrs, idStr)
				idSet[idStr] = true
			}
		}
	}

	filter := map[string]interface{}{
		"id": "in:" + strings.Join(idStrs, ","),
	}

	questionRepo.SetFilter(filter)
	questionRepo.SetLimit(len(questionIds)) // Set limit to exact number of questions needed
	questionRepo.SetPreload([]string{
		"Answers",
		"AnswerPositions",
		"AnswerGroups",
		"AnswerGroups.Group",
		"AnswerCoordinates",
		"AnswerMatchings",
		"RefAttributes",
		"RefAttributes.Attribute",
		"RefAttributes.ParentAttribute",
		"Source",
	})

	questions, err := questionRepo.GetAll()
	if err != nil {
		return err
	}

	questionIDMap := make(map[int64]*models.Question)
	for i := range questions {
		questionIDMap[questions[i].ID] = &questions[i]
	}

	var sortQuestionIds []int64
	sortQuestionIdsMap := make(map[int64]bool)

	if useOrder {
		// Build source_question_id -> questions (sorted by sort_position)
		sourceToQuestions := make(map[int64][]*models.Question)
		for i := range questions {
			q := questions[i]
			if q.SourceQuestionId > 0 {
				sourceToQuestions[q.SourceQuestionId] = append(sourceToQuestions[q.SourceQuestionId], &questions[i])
			}
		}
		for id, list := range sourceToQuestions {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].SortPosition < list[j].SortPosition
			})
			sourceToQuestions[id] = list
		}

		type orderItem struct {
			pos      int32
			isSource bool
			id       int64
		}
		var items []orderItem

		for idStr, pos := range req.QuestionsOrder {
			if idNew, err := strconv.Atoi(idStr); err == nil {
				id64 := int64(idNew)
				if _, ok := questionIDMap[id64]; ok {
					items = append(items, orderItem{pos: pos, isSource: false, id: id64})
				}
			}
		}
		for idStr, pos := range req.SourceQuestionsOrder {
			if idNew, err := strconv.Atoi(idStr); err == nil {
				items = append(items, orderItem{pos: pos, isSource: true, id: int64(idNew)})
			}
		}

		sort.SliceStable(items, func(i, j int) bool {
			return items[i].pos < items[j].pos
		})

		for _, it := range items {
			if !it.isSource {
				if !sortQuestionIdsMap[it.id] {
					sortQuestionIds = append(sortQuestionIds, it.id)
					sortQuestionIdsMap[it.id] = true
				}
			} else {
				if qs, ok := sourceToQuestions[it.id]; ok {
					for _, q := range qs {
						if !sortQuestionIdsMap[q.ID] {
							sortQuestionIds = append(sortQuestionIds, q.ID)
							sortQuestionIdsMap[q.ID] = true
						}
					}
				}
			}
		}

		// Append any remaining questions not mentioned in orders
		for _, q := range questions {
			if !sortQuestionIdsMap[q.ID] {
				sortQuestionIds = append(sortQuestionIds, q.ID)
				sortQuestionIdsMap[q.ID] = true
			}
		}
	} else {
		// Legacy sort based on sort_question_ids + scores
		for _, id := range req.SortQuestionIds {
			if _, ok := questionIDMap[id]; ok {
				sortQuestionIds = append(sortQuestionIds, id)
				sortQuestionIdsMap[id] = true
			}
		}

		scoreMap := make(map[int64]float64)
		isSort := true

		reqSortQuestionIdsMap := make(map[int64]bool)
		for _, id := range req.SortQuestionIds {
			reqSortQuestionIdsMap[id] = true
		}

		for sqIDStr, score := range req.SourceQuestionScores {
			var sqID int
			if sidNew, err := strconv.Atoi(sqIDStr); err == nil {
				sqID = sidNew
			}
			sqID64 := int64(sqID)

			for _, q := range questions {
				if q.SourceQuestionId == sqID64 {
					if existingScore, exists := scoreMap[q.ID]; exists {
						if existingScore != score {
							scoreMap[q.ID] = score
						}
					}

					if !sortQuestionIdsMap[q.ID] {
						if isSort || reqSortQuestionIdsMap[q.ID] {
							sortQuestionIds = append(sortQuestionIds, q.ID)
							sortQuestionIdsMap[q.ID] = true
						}
					}
				}
			}
		}

		for qIDStr, score := range req.QuestionScores {
			var qID int
			if idNew, err := strconv.Atoi(qIDStr); err == nil {
				qID = idNew
			}
			qID64 := int64(qID)

			if !sortQuestionIdsMap[qID64] {
				if isSort || reqSortQuestionIdsMap[qID64] {
					sortQuestionIds = append(sortQuestionIds, qID64)
					sortQuestionIdsMap[qID64] = true
				}
			}

			if existingScore, exists := scoreMap[qID64]; exists {
				if existingScore != score {
					scoreMap[qID64] = score
				}
			}
		}

		// Apply scores to questions
		for i := range questions {
			if score, ok := scoreMap[questions[i].ID]; ok && score > 0 {
				questions[i].Point = score
			}
		}
	}

	// Reorder questions slice to match sortQuestionIds before persisting.
	if len(sortQuestionIds) > 0 && len(questions) > 0 {
		idToIdx := make(map[int64]int, len(questions))
		for i := range questions {
			idToIdx[questions[i].ID] = i
		}

		ordered := make([]models.Question, 0, len(questions))
		used := make(map[int64]bool, len(sortQuestionIds))
		for _, id := range sortQuestionIds {
			if idx, ok := idToIdx[id]; ok && !used[id] {
				ordered = append(ordered, questions[idx])
				used[id] = true
			}
		}
		// Append remaining (safety)
		for _, q := range questions {
			if !used[q.ID] {
				ordered = append(ordered, q)
				used[q.ID] = true
			}
		}

		questions = ordered
	}

	if clonedQuestion.ID == 0 {
		var questionPtrs []*models.Question
		for i := range questions {
			if questions[i].ID == question.ID {
				questionPtrs = append(questionPtrs, question)
			} else {
				questionPtrs = append(questionPtrs, &questions[i])
			}
		}

		formatQuestions := questionResource.FormatQuestions(questionPtrs)

		for index, q := range formatQuestions {
			formatted := questionResource.FormatStripDomain(q)
			formatQuestions[index] = formatted
		}

		// Marshal questions to JSON
		marshalOptions := protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseProtoNames:   true,
		}

		var buf strings.Builder
		buf.WriteString("[")
		for i, q := range formatQuestions {
			b, err := marshalOptions.Marshal(q)
			if err != nil {
				config.Log.Errorf("marshal question failed: %v", err)
				continue
			}
			buf.Write(b)
			if i != len(formatQuestions)-1 {
				buf.WriteString(",")
			}
		}
		buf.WriteString("]")

		clonedQuestion.Questions = []byte(buf.String())

		// Marshal source_questions (unique per source_question_id, only metadata)
		sourceQuestionsJSON, err := marshalSourceQuestionsToJSONFromQuestions(formatQuestions)
		if err != nil {
			return err
		}

		// Marshal SortQuestionIds to JSON
		var sortQuestionIdsJSON datatypes.JSON
		if len(sortQuestionIds) > 0 {
			sortQuestionIdsBytes, err := json.Marshal(sortQuestionIds)
			if err != nil {
				return err
			}
			sortQuestionIdsJSON = datatypes.JSON(sortQuestionIdsBytes)
		} else {
			sortQuestionIdsJSON = datatypes.JSON([]byte("[]"))
		}

		cloned := models.ClonedQuestion{
			AssignmentID:          assignmentId,
			AssignmentType:        assignmentType,
			Questions:             []byte(buf.String()),
			SortQuestionIds:       sortQuestionIdsJSON,
			SortSourceQuestionIds: sortSourceQuestionIds,
			SourceQuestions:       datatypes.JSON(sourceQuestionsJSON),
		}

		if err := clonedRepo.Create(&cloned); err != nil {
			return err
		}
	} else {
		var existingRawQuestions []json.RawMessage
		if clonedQuestion.Questions != nil {
			if err := json.Unmarshal(clonedQuestion.Questions, &existingRawQuestions); err != nil {
				config.Log.Error("unmarshal clonedQuestion.Questions error:", err.Error())
				return err
			}
		}

		idMap := make(map[int]bool)
		for _, id := range questionIds {
			idMap[id] = true
		}

		list := make([]*prot.Question, 0, len(existingRawQuestions))
		for _, q := range existingRawQuestions {
			question, err := questionResource.ParseProtQuestionFromJSON(q)
			if err != nil {
				config.Log.Error("parse question error:", err.Error())
				return nil
			}
			if !idMap[int(question.Id)] || req.IsReset {
				continue
			}

			question = questionResource.FormatStripDomain(question)
			list = append(list, question)
		}

		existingMap := make(map[int64]bool)
		for _, q := range list {
			existingMap[q.Id] = true
		}

		for _, q := range questions {
			//if (q.QuestionType == models.QuestionTypeWriting || q.QuestionType == models.QuestionTypeSpeaking) && assignmentType == models.ClonedQuestionTypeHomework {
			//	continue
			//}

			if !existingMap[q.ID] {
				formatted := questionResource.FormatQuestion(&q)
				formatted = questionResource.FormatStripDomain(formatted)
				list = append(list, formatted)
			}
		}

		// Marshal questions to JSON
		marshalOptions := protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseProtoNames:   true,
		}

		var buf strings.Builder
		buf.WriteString("[")
		for i, q := range list {
			b, err := marshalOptions.Marshal(q)
			if err != nil {
				config.Log.Errorf("marshal question failed: %v", err)
				continue
			}
			buf.Write(b)
			if i != len(list)-1 {
				buf.WriteString(",")
			}
		}
		buf.WriteString("]")

		clonedQuestion.Questions = []byte(buf.String())

		// Marshal source_questions (unique per source_question_id, only metadata)
		sourceQuestionsJSON, err := marshalSourceQuestionsToJSONFromQuestions(list)
		if err != nil {
			return err
		}

		// Marshal SortQuestionIds to JSON
		if len(sortQuestionIds) > 0 {
			sortQuestionIdsBytes, err := json.Marshal(sortQuestionIds)
			if err != nil {
				return err
			}
			clonedQuestion.SortQuestionIds = datatypes.JSON(sortQuestionIdsBytes)
		} else {
			clonedQuestion.SortQuestionIds = datatypes.JSON([]byte("[]"))
		}

		// Update sort_source_question_ids with the latest order from request
		clonedQuestion.SortSourceQuestionIds = sortSourceQuestionIds
		clonedQuestion.SourceQuestions = datatypes.JSON(sourceQuestionsJSON)

		clonedRepo.SetContext(c)
		clonedRepo.Update(clonedQuestion)
	}

	// Sau khi cập nhật cloned question, cập nhật total_questions
	if assignmentType == models.ClonedQuestionTypeExam {
		_ = UpdateExamTotalQuestions(assignmentId)
	} else if assignmentType == models.ClonedQuestionTypeHomework {
		_ = UpdateHomeworkTotalQuestions(assignmentId)
	} else if assignmentType == models.ClonedQuestionTypeExercise {
		_ = UpdateExerciseTotalQuestions(assignmentId)
	}
	return err
}

// marshalSourceQuestionsToJSONFromQuestions builds a JSON array of unique source questions
// (one per source_question_id), without including the list of child questions.
func marshalSourceQuestionsToJSONFromQuestions(questions []*prot.Question) ([]byte, error) {
	marshalOptions := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}

	sourceMap := make(map[int64]*prot.SourceQuestion)
	for _, q := range questions {
		if q.Source != nil && q.Source.Id > 0 {
			if _, exists := sourceMap[q.Source.Id]; !exists {
				srcCopy := *q.Source
				sourceMap[q.Source.Id] = &srcCopy
			}
		}
	}

	if len(sourceMap) == 0 {
		return []byte("[]"), nil
	}

	var srcList []*prot.SourceQuestion
	for _, src := range sourceMap {
		srcList = append(srcList, src)
	}

	var buf strings.Builder
	buf.WriteString("[")
	for i, src := range srcList {
		b, err := marshalOptions.Marshal(src)
		if err != nil {
			config.Log.Errorf("marshal source question failed: %v", err)
			continue
		}
		buf.Write(b)
		if i != len(srcList)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")

	return []byte(buf.String()), nil
}
