package services

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
)

type QuestionService interface {
	GetAll(c *gin.Context) ([]models.Question, int64, error)
	GetCloned(c *gin.Context) ([]*prot.Question, int64, bool)
	GetByID(c *gin.Context, id int) (*prot.Question, error)
	GetClonedByID(c *gin.Context, id int, assignmentID int, assignmentType string) (*prot.Question, error)
	Create(c *gin.Context, req *prot.Question) (*models.Question, error)
	Update(c *gin.Context, req *prot.Question) (*models.Question, error)
	UpdateCloned(c *gin.Context, id int, assignmentID int, assignmentType string, req *prot.Question) (*prot.Question, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Question, error)
	Import(c *gin.Context, fileHeader *multipart.FileHeader) error
	Export(c *gin.Context) (string, error)
	ApplyFilter(c *gin.Context, filter map[string]interface{}) (map[string]interface{}, []repositories.QuesstionScore, error)
	GetQuestionIdAndKey(assignmentID int64, assignmentType string) ([]int64, string, error)
	GetKey(assignmentType string) (string, error)
	SyncKeywords(c *gin.Context, id int64) error
}

type questionService struct {
	repo  repositories.QuestionRepository
	meili MeiliService
}

func NewQuestionService(repo repositories.QuestionRepository) QuestionService {
	return &questionService{repo: repo, meili: NewMeiliService()}
}

func (s *questionService) GetAll(c *gin.Context) ([]models.Question, int64, error) {
	allowedFilters := []string{"status", "question_type", "subject_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	scores := []repositories.QuesstionScore{}
	filter, scores, _ = s.ApplyFilter(c, filter)

	// Enable base Meili integration for keyword ranking
	s.repo.SetMeili("questions", []string{"status", "question_type", "id"})
	s.repo.SetSearch(keyword, []string{"keywords", "id"})

	var questions []models.Question
	var rows int64
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Answers",
		"AnswerPositions",
		"AnswerGroups",
		"AnswerGroups.Group",
		"AnswerCoordinates",
		"AnswerMatchings",
		"RefAttributes",
		"RefAttributes.Attribute",
		"RefAttributes.ParentAttribute",
		"Subject",
	})

	questions, rows, err = s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	scoreMap := make(map[int64]float64)
	for _, s := range scores {
		scoreMap[s.QuestionID] = s.Score
	}

	for i := range questions {
		if score, ok := scoreMap[questions[i].ID]; ok && score > 0 {
			questions[i].Point = score
		}
	}

	return questions, rows, nil
}

func (s *questionService) GetByID(c *gin.Context, id int) (*prot.Question, error) {
	homeworkId, _ := strconv.Atoi(c.Query("homework_id"))
	examId, _ := strconv.Atoi(c.Query("exam_id"))
	lessonPlanPartId, _ := strconv.Atoi(c.Query("lesson_plan_part_id"))
	levelTestId, _ := strconv.Atoi(c.Query("level_test_id"))
	contestRoundId, _ := strconv.Atoi(c.Query("contest_round_id"))

	if homeworkId != 0 || examId != 0 || lessonPlanPartId != 0 || levelTestId != 0 || contestRoundId != 0 {
		var assignmentID int
		var assignmentType string

		switch {
		case homeworkId != 0:
			assignmentID = homeworkId
			assignmentType = models.ClonedQuestionTypeHomework
		case examId != 0:
			assignmentID = examId
			assignmentType = models.ClonedQuestionTypeExam
		case lessonPlanPartId != 0:
			assignmentID = lessonPlanPartId
			assignmentType = models.ClonedQuestionTypeLessonPlanPart
		case levelTestId != 0:
			assignmentID = levelTestId
			assignmentType = models.ClonedQuestionTypeLevelTest
		case contestRoundId != 0:
			assignmentID = contestRoundId
			assignmentType = models.ClonedQuestionTypeContestRound
		}

		return s.GetClonedByID(c, id, assignmentID, assignmentType)
	}

	s.repo.SetPreload([]string{
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

	question, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	questionResource := resources.NewQuestionResource()
	formattedQuestion := questionResource.FormatQuestion(question)

	return formattedQuestion, nil
}

func (s *questionService) Create(c *gin.Context, req *prot.Question) (*models.Question, error) {
	questionResource := resources.NewQuestionResource()
	question := questionResource.FormatModelQuestion(req)

	s.repo.SetContext(c)

	err := s.repo.Create(question)
	if err != nil {
		return nil, err
	}

	questionID := int64(question.ID)
	if err := s.StoreAnswers(questionID, req, req.Type); err != nil {
		return nil, err
	}

	if err := s.StoreAttribute(c, questionID, req); err != nil {
		return nil, err
	}
	if err := s.syncKeywordsFromRequest(questionID, req); err != nil {
		return nil, err
	}

	s.repo.SetPreload([]string{
		"Answers",
		"AnswerPositions",
		"AnswerGroups",
		"AnswerGroups.Group",
		"AnswerCoordinates",
		"AnswerMatchings",
		"RefAttributes",
		"RefAttributes.Attribute",
		"RefAttributes.ParentAttribute",
		"Subject",
	})

	newQuestion, _ := s.repo.FindNewByID(int(questionID))

	// Index to MeiliSearch (generic docs)
	if config.LoadConfig().MeiliEnabled && newQuestion != nil {
		doc := map[string]interface{}{
			"id":            newQuestion.ID,
			"status":        newQuestion.Status,
			"question_type": newQuestion.QuestionType,
			"title":         newQuestion.Title,
			"description":   newQuestion.Description,
			"content":       newQuestion.Content,
			"keywords":      newQuestion.Keywords,
			"point":         newQuestion.Point,
			"created_at":    newQuestion.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if err := s.meili.IndexDocuments("questions", []map[string]interface{}{doc}); err != nil {
			config.Log.Error("meili index create error:", err)
		}
	}

	return newQuestion, nil
}

func (s *questionService) Update(c *gin.Context, req *prot.Question) (*models.Question, error) {
	id, _ := strconv.Atoi(c.Param("id"))

	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
		"Answers",
		"AnswerPositions",
		"AnswerGroups",
		"AnswerGroups.Group",
		"AnswerCoordinates",
		"AnswerMatchings",
		"RefAttributes",
		"RefAttributes.Attribute",
		"RefAttributes.ParentAttribute",
		"Subject",
	})

	question, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	questionResource := resources.NewQuestionResource()
	updateQuestion := questionResource.FormatModelQuestion(req)

	updateQuestion.ID = question.ID
	if req.SubjectId == 0 {
		updateQuestion.SubjectId = question.SubjectId
	}

	err = s.repo.Update(updateQuestion)
	if err != nil {
		return nil, err
	}

	questionID := int64(question.ID)
	if err := s.StoreAnswers(questionID, req, req.Type); err != nil {
		return nil, err
	}

	if err := s.StoreAttribute(c, questionID, req); err != nil {
		return nil, err
	}
	if err := s.syncKeywordsFromRequest(questionID, req); err != nil {
		return nil, err
	}

	newQuestion, _ := s.repo.FindNewByID(id)

	// Re-index to MeiliSearch
	if config.LoadConfig().MeiliEnabled && newQuestion != nil {
		doc := map[string]interface{}{
			"id":            newQuestion.ID,
			"status":        newQuestion.Status,
			"question_type": newQuestion.QuestionType,
			"title":         newQuestion.Title,
			"description":   newQuestion.Description,
			"content":       newQuestion.Content,
			"keywords":      newQuestion.Keywords,
			"point":         newQuestion.Point,
			"created_at":    newQuestion.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if err := s.meili.IndexDocuments("questions", []map[string]interface{}{doc}); err != nil {
			config.Log.Error("meili index update error:", err)
		}
	}

	return newQuestion, nil
}

func (s *questionService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
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

	question, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// s.StoreCloned(c, question)

	switch question.QuestionType {
	case models.QuestionTypeMultipleChoice:
		if err := repositories.DeleteOldAnswers[models.Answer](int64(id), []int64{}); err != nil {
			return err
		}
	case models.QuestionTypeFillInBlanks, models.QuestionTypeOrdering, models.QuestionTypeDragDrop:
		if err := repositories.DeleteOldAnswers[models.AnswerPosition](int64(id), []int64{}); err != nil {
			return err
		}
	case models.QuestionTypeMatching:
		if err := repositories.DeleteOldAnswers[models.AnswerMatching](int64(id), []int64{}); err != nil {
			return err
		}
	case models.QuestionTypeLabeling:
		if err := repositories.DeleteOldAnswers[models.AnswerCoordinates](int64(id), []int64{}); err != nil {
			return err
		}
	case models.QuestionTypeCategory:
		if err := repositories.DeleteOldAnswers[models.AnswerGroup](int64(id), []int64{}); err != nil {
			return err
		}
	}
	s.repo.SetContext(c)

	err = s.repo.Delete(id)
	if err != nil {
		return err
	}

	// Delete from MeiliSearch
	if config.LoadConfig().MeiliEnabled {
		if err := s.meili.DeleteDocument("questions", question.ID); err != nil {
			config.Log.Error("meili delete error:", err)
		}
	}

	return nil
}

func (s *questionService) Restore(c *gin.Context, id int) (*models.Question, error) {
	s.repo.SetContext(c)
	question, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	// Re-index after restore
	if config.LoadConfig().MeiliEnabled && question != nil {
		doc := map[string]interface{}{
			"id":            question.ID,
			"status":        question.Status,
			"question_type": question.QuestionType,
			"title":         question.Title,
			"description":   question.Description,
			"content":       question.Content,
			"keywords":      question.Keywords,
			"point":         question.Point,
			"created_at":    question.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if err := s.meili.IndexDocuments("questions", []map[string]interface{}{doc}); err != nil {
			config.Log.Error("meili index restore error:", err)
		}
	}

	return question, nil
}

func (s *questionService) GetClonedByID(c *gin.Context, id int, assignmentID int, assignmentType string) (*prot.Question, error) {
	clonedRepo := repositories.NewClonedQuestionRepository()
	questionResource := resources.NewQuestionResource()
	clonedData, err := clonedRepo.FindByAssignment(int64(assignmentID), assignmentType)

	if errors.Is(err, gorm.ErrRecordNotFound) || clonedData.ID == 0 {
		return nil, err
	}

	var rawQuestions []json.RawMessage
	if err := json.Unmarshal(clonedData.Questions, &rawQuestions); err != nil {
		config.Log.Error("unmarshal rawQuestions error:", err.Error())
		return nil, err
	}

	for _, q := range rawQuestions {
		question, err := questionResource.ParseProtQuestionFromJSON(q)

		if err != nil {
			config.Log.Error(err)
			return nil, err
		}
		if question.Id == int64(id) {
			question = mergeCloneQuestionOptions(question)
			question = questionResource.FormatByRole(question, int64(utils.GetCurrentRoleId(c)))
			question = questionResource.FormatMediaUrls(question)
			return questionResource.FormatStaticURL(question), nil
		}
	}

	return nil, fmt.Errorf(i18n.Localize("messages.no_records_found"))
}

func (s *questionService) UpdateCloned(c *gin.Context, id int, assignmentID int, assignmentType string, req *prot.Question) (*prot.Question, error) {
	clonedRepo := repositories.NewClonedQuestionRepository()
	questionResource := resources.NewQuestionResource()
	questionRelationRepo := repositories.NewQuestionRelationRepository()
	clonedData, err := clonedRepo.FindByAssignment(int64(assignmentID), assignmentType)

	if errors.Is(err, gorm.ErrRecordNotFound) || clonedData.ID == 0 {
		return nil, err
	}

	isAssigned := false

	// Skip check assigned by homework
	if assignmentType == models.ClonedQuestionTypeHomework {
		// isAssigned, _ = questionRelationRepo.IsAssignedHomework(int64(assignmentID), 0)
	} else if assignmentType == models.ClonedQuestionTypeExam {
		isAssigned, _ = questionRelationRepo.IsAssignedExam(int64(assignmentID), 0)
	}

	config.Log.Info("isAssigned: ", isAssigned)

	if isAssigned {
		return nil, fmt.Errorf(i18n.Localize("messages.error_is_assigned"))
	}

	var rawQuestions []json.RawMessage
	if err := json.Unmarshal(clonedData.Questions, &rawQuestions); err != nil {
		config.Log.Error("unmarshal rawQuestions error:", err.Error())
		return nil, err
	}

	questions := make([]*prot.Question, 0, len(rawQuestions))

	for _, q := range rawQuestions {
		question, err := questionResource.ParseProtQuestionFromJSON(q)

		if err != nil {
			return nil, err
		}

		if question == nil {
			config.Log.Error("question is nil after parse, raw: %s", string(q))
			return nil, fmt.Errorf("parsed question is nil")
		}

		subjectRepo := repositories.NewSubjectRepository()

		if question.Id == int64(id) {
			req = questionResource.FormatStripDomain(req)
			if req.Id == 0 {
				req.Id = int64(id)
			}

			if req != nil && req.Options != nil {
				newAnswerId := uint64(1)
				for _, option := range req.Options.Answers {
					if option.Id > newAnswerId {
						newAnswerId = option.Id
					}
				}
				for i := range req.Options.Answers {
					if req.Options.Answers[i].Id == 0 {
						req.Options.Answers[i].Id = uint64(newAnswerId)
						newAnswerId += 1
					} else {
						newAnswerId = req.Options.Answers[i].Id + 1
					}
				}

				newLabelId := uint64(1)
				for _, option := range req.Options.Labels {
					if option.Id > newLabelId {
						newLabelId = option.Id
					}
				}
				for i := range req.Options.Labels {
					if req.Options.Labels[i].Id == 0 {
						req.Options.Labels[i].Id = uint64(newLabelId)
						newLabelId += 1
					} else {
						newLabelId = req.Options.Labels[i].Id + 1
					}
				}

				newBlankId := int64(1)
				for _, option := range req.Options.Blanks {
					if option.Id > newBlankId {
						newBlankId = option.Id
					}
				}
				for i := range req.Options.Blanks {
					if question.Type == models.QuestionTypeLabeling {
						for _, label := range req.Options.Labels {
							if label.CorrectPosition == int32(i+1) {
								req.Options.Blanks[i].Id = int64(label.Id)
								break
							}
						}
					} else {
						if req.Options.Blanks[i].Id == 0 {
							req.Options.Blanks[i].Id = int64(newBlankId)
							newBlankId += 1
						} else {
							newBlankId = req.Options.Blanks[i].Id + 1
						}
					}
				}

				newItemId := uint64(1)
				for _, option := range req.Options.Items {
					if option.Id > newItemId {
						newItemId = option.Id
					}
				}
				for i := range req.Options.Items {
					if req.Options.Items[i].Id == 0 {
						req.Options.Items[i].Id = uint64(newItemId)
						newItemId += 1
					} else {
						newItemId = req.Options.Items[i].Id + 1
					}
				}

				newCategoryId := uint64(1)
				for _, option := range req.Options.Categories {
					if option.Id > newCategoryId {
						newCategoryId = option.Id
					}
				}
				for i := range req.Options.Categories {
					if req.Options.Categories[i].Id == 0 {
						req.Options.Categories[i].Id = uint64(newCategoryId)
						newCategoryId += 1
					} else {
						newCategoryId = req.Options.Categories[i].Id + 1
					}
				}

				newSourceId := int64(1)
				for _, option := range req.Options.Sources {
					if option.Id > newSourceId {
						newSourceId = option.Id
					}
				}
				for i := range req.Options.Sources {
					if req.Options.Sources[i].Id == 0 {
						req.Options.Sources[i].Id = int64(newSourceId)
						newSourceId += 1
					} else {
						newSourceId = req.Options.Sources[i].Id + 1
					}
				}

				newTargetId := int64(1)
				for _, option := range req.Options.Targets {
					if option.Id > newTargetId {
						newTargetId = option.Id
					}
				}
				for i := range req.Options.Targets {
					if req.Options.Targets[i].Id == 0 {
						req.Options.Targets[i].Id = int64(newTargetId)
						newTargetId += 1
					} else {
						newTargetId = req.Options.Targets[i].Id + 1
					}
				}
			}

			req.CorrectAnswers = questionResource.FormatCorrectAnswers(req)
			if req.Options != nil {
				req.Options.Groups = questionResource.FormatGroups(req)
			}
			req = questionResource.FormatMedias(req)

			var subject *prot.QuestionSubjectInfo

			if req.SubjectId != 0 {
				subjectData, _ := subjectRepo.FindByID(int(req.SubjectId))

				if subjectData != nil {
					subject = &prot.QuestionSubjectInfo{
						Id:          subjectData.ID,
						Name:        subjectData.Name,
						Description: subjectData.Description,
					}
				}
			}

			req.Subject = subject

			questions = append(questions, req)
		} else {
			if err != nil {
				config.Log.Error("parse question error:", err.Error())
				return nil, err
			}
			questions = append(questions, question)
		}
	}

	questionsJSON, err := json.Marshal(questions)
	if err != nil {
		return nil, err
	}

	clonedData.Questions = questionsJSON

	clonedRepo.SetContext(c)
	err = clonedRepo.Update(clonedData)

	if err != nil {
		return nil, err
	}

	return questionResource.FormatStaticURL(req), nil
}

func (s *questionService) GetCloned(c *gin.Context) ([]*prot.Question, int64, bool) {
	clonedRepo := repositories.NewClonedQuestionRepository()
	questionResource := resources.NewQuestionResource()

	var assignmentID int64
	var assignmentType string
	var isRandom bool
	var hasCloned bool

	switch {
	case c.Query("homework_id") != "":
		id, err := strconv.ParseInt(c.Query("homework_id"), 10, 64)
		if err != nil {
			config.Log.Error("invalid homework_id:", err.Error())
			return nil, 0, hasCloned
		}
		assignmentID = id
		assignmentType = models.ClonedQuestionTypeHomework
		hasCloned = true
	case c.Query("exam_id") != "":
		id, err := strconv.ParseInt(c.Query("exam_id"), 10, 64)
		if err != nil {
			config.Log.Error("invalid exam_id:", err.Error())
			return nil, 0, hasCloned
		}
		assignmentID = id
		assignmentType = models.ClonedQuestionTypeExam
		hasCloned = true
	case c.Query("contest_round_id") != "":
		id, err := strconv.ParseInt(c.Query("contest_round_id"), 10, 64)
		if err != nil {
			config.Log.Error("invalid contest_round_id:", err.Error())
			return nil, 0, hasCloned
		}
		assignmentID = id
		assignmentType = models.ClonedQuestionTypeContestRound
		hasCloned = true
	case c.Query("exercise_id") != "":
		id, err := strconv.ParseInt(c.Query("exercise_id"), 10, 64)
		if err != nil {
			config.Log.Error("invalid exercise_id:", err.Error())
			return nil, 0, hasCloned
		}
		assignmentID = id
		assignmentType = models.ClonedQuestionTypeExercise
		hasCloned = true
	case c.Query("lesson_plan_part_id") != "":
		id, err := strconv.ParseInt(c.Query("lesson_plan_part_id"), 10, 64)
		if err != nil {
			config.Log.Error("invalid lesson_plan_part_id:", err.Error())
			return nil, 0, hasCloned
		}
		assignmentID = id
		assignmentType = models.ClonedQuestionTypeLessonPlanPart
		hasCloned = true
	case c.Query("level_test_id") != "":
		id, err := strconv.ParseInt(c.Query("level_test_id"), 10, 64)
		if err != nil {
			config.Log.Error("invalid lesson_plan_part_id:", err.Error())
			return nil, 0, hasCloned
		}
		assignmentID = id
		assignmentType = models.ClonedQuestionTypeLevelTest
		hasCloned = true
	case c.Query("contest_round_id") != "":
		id, err := strconv.ParseInt(c.Query("contest_round_id"), 10, 64)
		if err != nil {
			config.Log.Error("invalid contest_round_id:", err.Error())
			return nil, 0, hasCloned
		}
		assignmentID = id
		assignmentType = models.ClonedQuestionTypeContestRound
		hasCloned = true
		config.Log.Info(fmt.Sprintf("🔍 Contest round ID: %d, AssignmentType: %s, HasCloned: %t", assignmentID, assignmentType, hasCloned))
	default:
		return nil, 0, hasCloned
	}

	questions := []*prot.Question{}
	cloned, err := clonedRepo.FindByAssignment(assignmentID, assignmentType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config.Log.Info(fmt.Sprintf("🔍 No cloned questions found for assignmentID: %d, assignmentType: %s", assignmentID, assignmentType))
			return questions, 0, hasCloned
		}
		config.Log.Error("find cloned error:", err.Error())
		return questions, 0, hasCloned
	}
	config.Log.Info(fmt.Sprintf("🔍 Found cloned questions for assignmentID: %d, assignmentType: %s", assignmentID, assignmentType))

	var rawQuestions []json.RawMessage
	if err := json.Unmarshal(cloned.Questions, &rawQuestions); err != nil {
		config.Log.Error("unmarshal rawQuestions error:", err.Error())
		return nil, 0, hasCloned
	}

	healedClone := false
	mergedForPersist := make([]*prot.Question, 0, len(rawQuestions))
	roleId := int64(utils.GetCurrentRoleId(c))

	for _, q := range rawQuestions {
		question, err := questionResource.ParseProtQuestionFromJSON(q)
		if err != nil {
			config.Log.Error("parse question error:", err.Error())
			return questions, 0, hasCloned
		}

		needsHeal := clonedQuestionNeedsOptionsRefresh(question)
		question = mergeCloneQuestionOptions(question)
		if needsHeal && !clonedQuestionNeedsOptionsRefresh(question) {
			healedClone = true
		}
		mergedForPersist = append(mergedForPersist, cloneProtQuestion(question))

		question = questionResource.FormatByRole(question, roleId)
		question = questionResource.FormatStaticURL(question)
		question = questionResource.FormatMediaUrls(question)

		questions = append(questions, question)
	}

	if healedClone {
		if err := persistCloneQuestionsSnapshot(assignmentID, assignmentType, mergedForPersist); err != nil {
			config.Log.Errorf("persist healed clone snapshot: %v", err)
		}
	}

	if roleId == models.StudentRoleId {
		switch assignmentType {
		case models.ClonedQuestionTypeHomework:
			homeworkRepo := repositories.NewHomeworkRepository()
			userId := int64(utils.GetCurrentUserId(c))
			homework, err := homeworkRepo.GetByID(int64(assignmentID), userId, c)

			if err == nil && homework.ID > 0 {
				if homework.IsRandomQuestion {
					isRandom = true
				}
			}
		case models.ClonedQuestionTypeExam:
			examRepo := repositories.NewExamRepository()
			exam, _ := examRepo.GetByID(int64(assignmentID), c)

			if err == nil && exam.ID > 0 {
				if exam.IsRandomQuestion {
					isRandom = true
				}
			}
		case models.ClonedQuestionTypeExercise:
			exerciseRepo := repositories.NewExerciseRepository()
			exercise, _ := exerciseRepo.GetByID(int64(assignmentID), c)

			if err == nil && exercise.ID > 0 {
				if exercise.IsRandomQuestion {
					isRandom = true
				}
			}
		}
	}

	if isRandom {
		rand.Shuffle(len(questions), func(i, j int) {
			questions[i], questions[j] = questions[j], questions[i]
		})
	}

	return questions, int64(len(questions)), hasCloned
}

func (s *questionService) SyncKeywords(c *gin.Context, id int64) error {
	var questions []models.Question

	// Nếu id = 0 thì sync tất cả
	if id == 0 {
		if err := db.MasterDB.Preload("Answers").
			Preload("AnswerPositions").
			Preload("AnswerGroups").
			Preload("AnswerCoordinates").
			Preload("AnswerMatchings").
			Where("keywords IS NULL OR keywords = ''").
			Find(&questions).Error; err != nil {
			return err
		}
	} else {
		var question models.Question
		if err := db.MasterDB.Preload("Answers").
			Preload("AnswerPositions").
			Preload("AnswerGroups").
			Preload("AnswerCoordinates").
			Preload("AnswerMatchings").
			First(&question, id).Error; err != nil {
			return err
		}
		questions = append(questions, question)
	}

	// Lặp tất cả questions để sync Keywords
	for _, q := range questions {
		var keywords []string

		if q.Title != "" {
			keywords = append(keywords, q.Title)
		}
		if q.Description != "" {
			keywords = append(keywords, q.Description)
		}
		if q.Content != "" {
			keywords = append(keywords, q.Content)
		}

		for _, ans := range q.Answers {
			if ans.Content != "" {
				keywords = append(keywords, ans.Content)
			}
		}
		for _, pos := range q.AnswerPositions {
			if pos.Content != "" {
				keywords = append(keywords, pos.Content)
			}
		}
		for _, grp := range q.AnswerGroups {
			if grp.Content != "" {
				keywords = append(keywords, grp.Content)
			}
		}
		for _, grp := range q.AnswerCoordinates {
			if grp.Content != "" {
				keywords = append(keywords, grp.Content)
			}
		}
		for _, grp := range q.AnswerMatchings {
			if grp.Content != "" {
				keywords = append(keywords, grp.Content)
			}
		}

		newKeywords := strings.Join(keywords, " | ")

		// Cập nhật chỉ trường Keywords
		if err := db.MasterDB.Model(&q).Update("keywords", newKeywords).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *questionService) GetQuestionIdAndKey(assignmentID int64, assignmentType string) ([]int64, string, error) {
	var questionIDs []int64
	var filterKey string

	switch assignmentType {
	case models.ClonedQuestionTypeHomework:
		// Không còn bảng homework_questions, trả về slice rỗng
		records := []repositories.QuesstionScore{}
		for _, record := range records {
			questionIDs = append(questionIDs, record.QuestionID)
		}
		filterKey = "homework_id"

	case models.ClonedQuestionTypeExam:
		// Không còn bảng exam_questions, trả về slice rỗng
		records := []repositories.QuesstionScore{}
		for _, record := range records {
			questionIDs = append(questionIDs, record.QuestionID)
		}
		filterKey = "exam_id"

	case models.ClonedQuestionTypeLessonPlanPart:
		records, err := s.repo.GetQuestionIdsAndScoresByLessonPlanPartId(assignmentID)
		if err != nil {
			return nil, "", err
		}
		for _, record := range records {
			questionIDs = append(questionIDs, record.QuestionID)
		}
		filterKey = "lesson_plan_part_id"

	case models.ClonedQuestionTypeLevelTest:
		records, err := s.repo.GetQuestionIdsAndScoresByLevelTestId(assignmentID)
		if err != nil {
			return nil, "", err
		}
		for _, record := range records {
			questionIDs = append(questionIDs, record.QuestionID)
		}
		filterKey = "level_test_id"

	case models.ClonedQuestionTypeExercise:
		// Không còn bảng exercise_questions, trả về slice rỗng
		records := []repositories.QuesstionScore{}
		for _, record := range records {
			questionIDs = append(questionIDs, record.QuestionID)
		}
		filterKey = "exercise_id"

	case models.ClonedQuestionTypeContestRound:
		// For contest rounds, we use cloned_questions table
		clonedRepo := repositories.NewClonedQuestionRepository()
		cloned, err := clonedRepo.FindByAssignment(assignmentID, assignmentType)
		if err != nil {
			return nil, "", err
		}
		// Parse questions from JSON to get question IDs
		var questions []map[string]interface{}
		if err := json.Unmarshal(cloned.Questions, &questions); err != nil {
			return nil, "", err
		}
		for _, q := range questions {
			if id, ok := q["id"].(float64); ok {
				questionIDs = append(questionIDs, int64(id))
			}
		}
		filterKey = "contest_round_id"

	default:
		return nil, "", fmt.Errorf("unsupported assignment type: %s", assignmentType)
	}

	return questionIDs, filterKey, nil
}

func clonedQuestionNeedsOptionsRefresh(question *prot.Question) bool {
	if question == nil {
		return false
	}

	opts := question.Options
	switch question.Type {
	case models.QuestionTypeMultipleChoice, models.QuestionTypeFillInBlanks,
		models.QuestionTypeOrdering, models.QuestionTypeDragDrop:
		return opts == nil || len(opts.Answers) == 0
	case models.QuestionTypeMatching:
		return opts == nil || len(opts.Sources) == 0 || len(opts.Targets) == 0
	case models.QuestionTypeCategory:
		return opts == nil || len(opts.Categories) == 0 || len(opts.Items) == 0
	case models.QuestionTypeLabeling:
		return opts == nil || len(opts.Labels) == 0
	default:
		return false
	}
}

func refreshQuestionOptionsFromDB(questionID int64) (*prot.Question, error) {
	questionRepo := repositories.NewQuestionRepository()
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

	question, err := questionRepo.FindByID(int(questionID))
	if err != nil {
		return nil, err
	}

	questionResource := resources.NewQuestionResource()
	return questionResource.FormatQuestion(question), nil
}

// mergeCloneQuestionOptions prefers live DB options when the clone snapshot is empty
// or when the question is category (options are stored in answer_groups / group_answers).
func mergeCloneQuestionOptions(question *prot.Question) *prot.Question {
	if question == nil {
		return question
	}

	cloneMissing := clonedQuestionNeedsOptionsRefresh(question)
	preferDB := question.Type == models.QuestionTypeCategory
	if !cloneMissing && !preferDB {
		return question
	}

	refreshed, err := refreshQuestionOptionsFromDB(question.Id)
	if err != nil || refreshed == nil || refreshed.Options == nil {
		return question
	}

	dbUsable := !clonedQuestionNeedsOptionsRefresh(refreshed)
	if !dbUsable {
		return question
	}

	question.Options = refreshed.Options
	if refreshed.CorrectAnswers != nil {
		question.CorrectAnswers = refreshed.CorrectAnswers
	}

	return question
}

// categoryQuestionForScoring loads live DB options/answers for category scoring.
// Clone snapshots are often missing or stale after question-bank edits.
func categoryQuestionForScoring(cloned repositories.ClonedQuestion) (*prot.Question, error) {
	questionID, err := strconv.ParseInt(cloned.ID.String(), 10, 64)
	if err != nil {
		return nil, err
	}

	if refreshed, refreshErr := refreshQuestionOptionsFromDB(questionID); refreshErr == nil && refreshed != nil {
		if !clonedQuestionNeedsOptionsRefresh(refreshed) {
			refreshed.Id = questionID
			if cloned.Type != "" {
				refreshed.Type = cloned.Type
			}
			return refreshed, nil
		}
	}

	question := &prot.Question{
		Id:   questionID,
		Type: cloned.Type,
	}

	unmarshalOpts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if len(cloned.Options) > 0 {
		question.Options = &prot.QuestionOption{}
		if err := unmarshalOpts.Unmarshal(cloned.Options, question.Options); err != nil {
			if err := parseCategoryOptionsFromCloneJSON(cloned.Options, question); err != nil {
				return nil, fmt.Errorf("parse cloned options: %w", err)
			}
		}
	}
	if len(cloned.CorrectAnswers) > 0 {
		question.CorrectAnswers = &prot.QuestionCorrectAnswers{}
		if err := unmarshalOpts.Unmarshal(cloned.CorrectAnswers, question.CorrectAnswers); err != nil {
			if err := parseCategoryCorrectAnswersFromCloneJSON(cloned.CorrectAnswers, question); err != nil {
				return nil, fmt.Errorf("parse cloned correct_answers: %w", err)
			}
		}
	}

	return mergeCloneQuestionOptions(question), nil
}

func parseCategoryOptionsFromCloneJSON(raw json.RawMessage, question *prot.Question) error {
	var payload struct {
		Items []struct {
			ID            json.Number `json:"id"`
			Text          string      `json:"text"`
			Point         float64     `json:"point"`
			GroupPosition int64       `json:"group_position"`
		} `json:"items"`
		Categories []struct {
			ID   json.Number `json:"id"`
			Name string      `json:"name"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	question.Options = &prot.QuestionOption{}
	for _, cat := range payload.Categories {
		id, _ := strconv.ParseUint(cat.ID.String(), 10, 64)
		question.Options.Categories = append(question.Options.Categories, &prot.GroupAnswer{
			Id:   id,
			Name: cat.Name,
		})
	}
	for _, item := range payload.Items {
		id, _ := strconv.ParseUint(item.ID.String(), 10, 64)
		question.Options.Items = append(question.Options.Items, &prot.AnswerContent{
			Id:            id,
			Text:          item.Text,
			Point:         item.Point,
			GroupPosition: int32(item.GroupPosition),
		})
	}
	return nil
}

func parseCategoryCorrectAnswersFromCloneJSON(raw json.RawMessage, question *prot.Question) error {
	var payload struct {
		List map[string]string `json:"list"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	question.CorrectAnswers = &prot.QuestionCorrectAnswers{List: payload.List}
	return nil
}

func categoryCorrectGroupID(question *prot.Question, answerID int64, answerIDStr string) int64 {
	if question == nil {
		return 0
	}

	if question.CorrectAnswers != nil && question.CorrectAnswers.List != nil {
		if v, ok := question.CorrectAnswers.List[answerIDStr]; ok {
			if correctGroupID, err := strconv.ParseInt(v, 10, 64); err == nil && correctGroupID > 0 {
				return correctGroupID
			}
		}
	}

	if question.Options == nil {
		return 0
	}

	for _, item := range question.Options.Items {
		if int64(item.Id) != answerID || item.GroupPosition <= 0 {
			continue
		}
		pos := int(item.GroupPosition)
		for i, cat := range question.Options.Categories {
			if i+1 == pos {
				return int64(cat.Id)
			}
		}
	}

	return 0
}

func buildCategoryGroupScoreResults(
	reqAnswers map[string]int64,
	question *prot.Question,
	perGroupScore float64,
) (totalScore float64, correctCount int, groupResults []*prot.GroupPair) {
	itemIdToText := make(map[int64]string)
	groupIdToName := make(map[int64]string)
	if question != nil && question.Options != nil {
		for _, item := range question.Options.Items {
			itemIdToText[int64(item.Id)] = item.Text
		}
		for _, cat := range question.Options.Categories {
			groupIdToName[int64(cat.Id)] = cat.Name
		}
	}

	for answerIDStr, groupID := range reqAnswers {
		answerID, err := strconv.ParseInt(answerIDStr, 10, 64)
		if err != nil {
			continue
		}

		correctGroupID := categoryCorrectGroupID(question, answerID, answerIDStr)
		isCorrect := correctGroupID > 0 && groupID == correctGroupID
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

	return totalScore, correctCount, groupResults
}

// EnrichClonedQuestionMaps fills empty category options/correct_answers from live DB
// for homework-answers and other APIs that read cloned_questions JSON directly.
func EnrichClonedQuestionMaps(questions []map[string]interface{}) []map[string]interface{} {
	if len(questions) == 0 {
		return questions
	}

	questionResource := resources.NewQuestionResource()
	marshalOpts := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}

	for i, rawMap := range questions {
		raw, err := json.Marshal(rawMap)
		if err != nil {
			continue
		}
		q, err := questionResource.ParseProtQuestionFromJSON(raw)
		if err != nil {
			continue
		}
		merged := mergeCloneQuestionOptions(q)
		b, err := marshalOpts.Marshal(merged)
		if err != nil {
			continue
		}
		var updated map[string]interface{}
		if err := json.Unmarshal(b, &updated); err != nil {
			continue
		}
		questions[i] = updated
	}

	return questions
}

func cloneProtQuestion(q *prot.Question) *prot.Question {
	if q == nil {
		return nil
	}
	b, err := protojson.Marshal(q)
	if err != nil {
		return q
	}
	var out prot.Question
	if err := protojson.Unmarshal(b, &out); err != nil {
		return q
	}
	return &out
}

func persistCloneQuestionsSnapshot(assignmentID int64, assignmentType string, questions []*prot.Question) error {
	questionResource := resources.NewQuestionResource()
	marshalOpts := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}

	var buf strings.Builder
	buf.WriteString("[")
	for i, q := range questions {
		stripped := questionResource.FormatStripDomain(q)
		b, err := marshalOpts.Marshal(stripped)
		if err != nil {
			return err
		}
		buf.Write(b)
		if i != len(questions)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")

	return db.MasterDB.Model(&models.ClonedQuestion{}).
		Where("assignment_id = ? AND assignment_type = ?", assignmentID, assignmentType).
		Update("questions", []byte(buf.String())).Error
}
