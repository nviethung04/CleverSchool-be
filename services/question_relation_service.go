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
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
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
	fmt.Printf("🔍 AssignQuestions called: ExamId=%d, HomeworkId=%d, ContestRoundId=%d, LevelTestId=%d, LessonPlanPartId=%d, ExerciseId=%d, QuestionScores=%v, IsReset=%v\n",
		req.ExamId, req.HomeworkId, req.ContestRoundId, req.LevelTestId, req.LessonPlanPartId, req.ExerciseId, req.QuestionScores, req.IsReset)

	if len(req.QuestionScores) == 0 && len(req.SourceQuestionScores) == 0 && !req.IsReset && !req.ClearSelfScoringQuestions {
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
			fmt.Printf("🔍 ResetAll with ContestRoundId=%d\n", req.ContestRoundId)
			return s.UpdateCloneQuestion(c.Copy(), int64(req.ContestRoundId), models.ClonedQuestionTypeContestRound, req)
		}
		if req.HomeworkId > 0 {
			fmt.Printf("🔍 ResetAll with HomeworkId=%d\n", req.HomeworkId)
			return s.UpdateCloneQuestion(c.Copy(), int64(req.HomeworkId), models.ClonedQuestionTypeHomework, req)
		}
		if req.ExamId > 0 {
			fmt.Printf("🔍 ResetAll with ExamId=%d\n", req.ExamId)
			return s.UpdateCloneQuestion(c.Copy(), int64(req.ExamId), models.ClonedQuestionTypeExam, req)
		}
		if req.LevelTestId > 0 {
			fmt.Printf("🔍 ResetAll with LevelTestId=%d\n", req.LevelTestId)
			return s.UpdateCloneQuestion(c.Copy(), int64(req.LevelTestId), models.ClonedQuestionTypeLevelTest, req)
		}
		if req.LessonPlanPartId > 0 {
			fmt.Printf("🔍 ResetAll with LessonPlanPartId=%d\n", req.LessonPlanPartId)
			return s.UpdateCloneQuestion(c.Copy(), int64(req.LessonPlanPartId), models.ClonedQuestionTypeLessonPlanPart, req)
		}
		if req.ExerciseId > 0 {
			fmt.Printf("🔍 ResetAll with ExerciseId=%d\n", req.ExerciseId)
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
		if len(req.QuestionScores) > 0 || req.IsReset {
			fmt.Printf("🔍 Entering main condition: QuestionScores=%v, IsReset=%v\n", req.QuestionScores, req.IsReset)
			if req.ExamId > 0 {
				fmt.Printf("🔍 Processing ExamId=%d\n", req.ExamId)
				if !isAssignedExam || req.IsReset {
					return s.UpdateCloneQuestion(c.Copy(), int64(req.ExamId), models.ClonedQuestionTypeExam, req)
				} else {
					return fmt.Errorf(i18n.Localize("messages.error_is_assigned"))
				}
			}
			if req.ExerciseId > 0 {
				fmt.Printf("🔍 Processing ExerciseId=%d\n", req.ExerciseId)
				return s.UpdateCloneQuestion(c.Copy(), int64(req.ExerciseId), models.ClonedQuestionTypeExercise, req)
			}
			if req.LessonPlanPartId > 0 {
				fmt.Printf("🔍 Processing LessonPlanPartId=%d\n", req.LessonPlanPartId)
				return s.UpdateCloneQuestion(c.Copy(), int64(req.LessonPlanPartId), models.ClonedQuestionTypeLessonPlanPart, req)
			}
			if req.HomeworkId > 0 {
				fmt.Printf("🔍 Processing HomeworkId=%d\n", req.HomeworkId)
				// Skip check assigned by homework
				if !isAssignedHomework || req.IsReset {
					return s.UpdateCloneQuestion(c.Copy(), int64(req.HomeworkId), models.ClonedQuestionTypeHomework, req)
				} else {
					return fmt.Errorf(i18n.Localize("messages.error_is_assigned"))
				}
			}
			if req.LevelTestId > 0 {
				fmt.Printf("🔍 Processing LevelTestId=%d\n", req.LevelTestId)
				return s.UpdateCloneQuestion(c.Copy(), int64(req.LevelTestId), models.ClonedQuestionTypeLevelTest, req)
			}
			if req.ContestRoundId > 0 {
				fmt.Printf("🔍 QuestionRelationService: Processing contest_round_id=%d\n", req.ContestRoundId)
				return s.UpdateCloneQuestion(c.Copy(), int64(req.ContestRoundId), models.ClonedQuestionTypeContestRound, req)
			}
			fmt.Printf("🔍 No valid assignment ID found!\n")
		} else {
			fmt.Printf("🔍 Skipping main condition: QuestionScores=%v, IsReset=%v\n", req.QuestionScores, req.IsReset)
		}
	}

	return err
}

func (s *questionRelationService) UpdateCloneQuestion(c *gin.Context, assignmentId int64, assignmentType string, req *prot.CreateQuestionRelationRequest) error {
	fmt.Printf("🔍 UpdateCloneQuestion: assignmentId=%d, assignmentType=%s\n", assignmentId, assignmentType)
	clonedRepo := repositories.NewClonedQuestionRepository()
	questionRepo := repositories.NewQuestionRepository()
	questionService := NewQuestionService(questionRepo)
	clonedData, err := clonedRepo.FindByAssignment(assignmentId, assignmentType)
	filterKey, _ := questionService.GetKey(assignmentType)

	if errors.Is(err, gorm.ErrRecordNotFound) || clonedData.ID == 0 {
		fmt.Printf("🔍 CreateOrUpdateCloneQuestion: Creating new (assignmentId=%d, assignmentType=%s)\n", assignmentId, assignmentType)
		return s.CreateOrUpdateCloneQuestion(c, assignmentId, assignmentType, filterKey, &models.Question{}, req, &models.ClonedQuestion{})
	} else {
		fmt.Printf("🔍 CreateOrUpdateCloneQuestion: Updating existing (assignmentId=%d, assignmentType=%s, clonedData.ID=%d)\n", assignmentId, assignmentType, clonedData.ID)
		return s.CreateOrUpdateCloneQuestion(c, assignmentId, assignmentType, filterKey, &models.Question{}, req, clonedData)
	}
}

func (s *questionRelationService) CreateOrUpdateCloneQuestion(c *gin.Context, assignmentId int64, assignmentType string, filterKey string, question *models.Question, req *prot.CreateQuestionRelationRequest, clonedQuestion *models.ClonedQuestion) error {
	questionResource := resources.NewQuestionResource()
	clonedRepo := repositories.NewClonedQuestionRepository()
	clonedRepo.SetContext(c)

	questionRepo := repositories.NewQuestionRepository()

	questionIds := make([]int, 0, len(req.QuestionScores))
	fmt.Printf("🔍 CreateOrUpdateCloneQuestion: assignmentId=%d, assignmentType=%s, clonedQuestion.ID=%d, QuestionScores=%v\n", assignmentId, assignmentType, clonedQuestion.ID, req.QuestionScores)

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
	} else {
		for idStr := range req.QuestionScores {
			var id int
			if idNew, err := strconv.Atoi(idStr); err == nil {
				id = idNew
			}
			questionIds = append(questionIds, id)
		}
	}

	idStrs := make([]string, len(questionIds))

	for i, id := range questionIds {
		idStrs[i] = strconv.Itoa(id)
	}

	filter := map[string]interface{}{
		"id": "in:" + strings.Join(idStrs, ","),
	}

	// Debug log
	config.Log.Infof("Question IDs to filter: %v", questionIds)
	config.Log.Infof("Filter: %v", filter)

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
	})

	questions, _, err := questionRepo.FindAll()
	if err != nil {
		return err
	}

	// Debug log
	config.Log.Infof("Found %d questions for IDs %v", len(questions), questionIds)
	for _, q := range questions {
		config.Log.Infof("Question ID: %d, Title: %s", q.ID, q.Title)
	}

	var questionScores []repositories.QuesstionScore

	scoreMap := make(map[int64]float64)

	for qIDStr, score := range req.QuestionScores {
		var qID int
		if idNew, err := strconv.Atoi(qIDStr); err == nil {
			qID = idNew
		}
		qID64 := int64(qID)
		if existingScore, exists := scoreMap[qID64]; exists {
			if existingScore != score {
				scoreMap[qID64] = score
			}
		} else {
			questionScores = append(questionScores, repositories.QuesstionScore{
				QuestionID: qID64,
				Score:      score,
			})
		}
	}

	if clonedQuestion.ID == 0 {
		for i := range questions {
			if score, ok := scoreMap[questions[i].ID]; ok && score > 0 {
				questions[i].Point = score
			}
		}

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

		cloned := models.ClonedQuestion{
			AssignmentID:   assignmentId,
			AssignmentType: assignmentType,
			Questions:      []byte(buf.String()),
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
